package bot

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mischief/glenda/util"

	"github.com/kballard/goirc/irc"
	"github.com/koding/kite"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type HookFn func(b *Bot, sender, cmd string, args ...string) error

type BotConfig struct {
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	SSL      bool     `json:"ssl"`
	Nick     string   `json:"nick"`
	User     string   `json:"user"`
	Real     string   `json:"real"`
	Channels []string `json:"channels"`
}

type Bot struct {
	viper  *viper.Viper
	kite   *kite.Kite
	config *BotConfig

	conn      irc.SafeConn
	ircConfig irc.Config

	quit chan bool

	// map of kite ids to clients that are notified on privmsgs.
	notifyPrivmsg map[string]*kite.Client
	// notifyMu protects notifyPrivmsg
	notifyMu sync.Mutex
}

func NewBot(etcd, conf string) (*Bot, error) {
	var config BotConfig

	viper := viper.New()
	viper.AddRemoteProvider("etcd", etcd, conf)
	viper.SetConfigType("json")

	if err := viper.ReadRemoteConfig(); err != nil {
		return nil, fmt.Errorf("failed to retrieve config: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	k := util.NewKite("irc", "1.0.0")

	bot := &Bot{
		viper:         viper,
		kite:          k,
		config:        &config,
		quit:          make(chan bool, 1),
		notifyPrivmsg: make(map[string]*kite.Client),
	}

	bot.ircConfig = irc.Config{
		Host: config.Host,
		Port: uint(config.Port),
		SSL:  config.SSL,
		// XXX(mischief): fix me
		SSLConfig:  &tls.Config{InsecureSkipVerify: true},
		Nick:       config.Nick,
		User:       config.User,
		RealName:   config.Real,
		Timeout:    10 * time.Second,
		AllowFlood: true,
		Init:       bot.init,
	}

	k.HandleFunc("irc.config", bot.configHandler).DisableAuthentication()
	k.HandleFunc("irc.register", bot.registerHandler).DisableAuthentication()

	k.HandleFunc("irc.privmsg", bot.privmsgHandler).DisableAuthentication()
	k.HandleFunc("irc.join", bot.joingHandler).DisableAuthentication()
	k.HandleFunc("irc.part", bot.partHandler).DisableAuthentication()

	return bot, nil
}

func (b *Bot) init(hr irc.HandlerRegistry) {
	hr.AddHandler(irc.CONNECTED, b.connect)
	hr.AddHandler(irc.DISCONNECTED, b.disconnect)
	hr.AddHandler("PRIVMSG", b.privmsg)
	hr.AddHandler(irc.ACTION, b.action)
}

func (b *Bot) connect(conn *irc.Conn, l irc.Line) {
	b.kite.Log.Info("connected")
	for _, c := range b.config.Channels {
		b.kite.Log.Info("joining %q", c)
		b.conn.Join([]string{c}, nil)
	}
}

func (b *Bot) disconnect(conn *irc.Conn, l irc.Line) {
	b.kite.Log.Warning("disconnected")
	b.quit <- true
}

func (b *Bot) privmsg(conn *irc.Conn, l irc.Line) {
	b.kite.Log.Info("[%s] %s> %s\n", l.Args[0], l.Src, l.Args[1])

	args := strings.Split(l.Args[1], " ")
	cmd := args[0]

	args = args[1:]

	if cmd == "" {
		return
	}

	sender := l.Src.String()

	if l.Args[0][0] == '#' {
		sender = l.Args[0]
	}

	// broadcast to kites
	out := map[string]string{
		"sender":   l.Src.String(),
		"reply-to": sender,
		"message":  l.Args[1],
	}

	b.notifyMu.Lock()
	for _, k := range b.notifyPrivmsg {
		_, err := k.Tell("irc.privmsg", out)
		if err != nil {
			b.kite.Log.Warning("ircmessage callback error to kite %v: %v", k, err)
		}
	}
	b.notifyMu.Unlock()
}

func (b *Bot) action(conn *irc.Conn, l irc.Line) {
	b.kite.Log.Info("[%s] %s %s", l.Dst, l.Src, l.Args[0])
}

func (b *Bot) configHandler(r *kite.Request) (result interface{}, err error) {
	return b.config, nil
}

func (b *Bot) registerHandler(r *kite.Request) (result interface{}, err error) {
	b.notifyMu.Lock()
	defer b.notifyMu.Unlock()
	b.notifyPrivmsg[r.Client.ID] = r.Client

	var once sync.Once
	r.Client.OnDisconnect(func() {
		once.Do(func() {
			b.notifyMu.Lock()
			defer b.notifyMu.Unlock()
			delete(b.notifyPrivmsg, r.Client.ID)
		})
	})

	return nil, nil
}

func (b *Bot) privmsgHandler(r *kite.Request) (result interface{}, err error) {
	var args map[string]string
	r.Args.One().MustUnmarshal(&args)
	b.conn.Privmsg(args["reply-to"], args["message"])
	return nil, nil
}

func (b *Bot) joinHandler(r *kite.Request) (result interface{}, err error) {
	// "#channel": "key"
	var args map[string]string
	r.Args.One().MustUnmarshal(&args)

	for c, k := range args {
		b.kite.Log.Info("joining %q", c)
		var key []string
		if k != "" {
			key = []string{k}
		}
		b.conn.Join([]string{c}, key)
	}

	return nil, nil
}

func (b *Bot) partHandler(r *kite.Request) (result interface{}, err error) {
	var args struct {
		channels []string
		reason   string
	}

	r.Args.One().MustUnmarshal(&args)
	b.conn.Part(args.channels, args.reason)

	return nil, nil
}

func (b *Bot) Run() {
	b.kite.Log.Info("connecting to %s:%d...", b.config.Host, b.config.Port)
	conn, err := irc.Connect(b.ircConfig)
	if err != nil {
		close(b.quit)
		b.kite.Log.Fatal("error connecting: %v", err)
	}

	b.conn = conn

	<-b.quit
}
