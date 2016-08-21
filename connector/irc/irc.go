// +build irc

package irc

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/fluffle/goirc/client"
	"github.com/fluffle/goirc/state"
	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/mischief/glenda/connector"
	"github.com/mischief/glenda/core"
)

func init() {
	connector.Register("irc", New)
}

var ircOptions = []string{"host", "nick", "ident", "channels", "tls"}

func ircConf(conf *viper.Viper) (host, nick, ident string, channels []string, tls bool, err error) {
	for _, o := range ircOptions {
		if !conf.IsSet(o) {
			err = fmt.Errorf("missing irc option %q", o)
			return
		}
	}

	host = conf.GetString("host")
	nick = conf.GetString("nick")
	ident = conf.GetString("ident")
	channels = conf.GetStringSlice("channels")
	tls = conf.GetBool("tls")
	return
}

type irc struct {
	mu         sync.Mutex
	conn       *client.Conn
	reconnect  bool
	quit       chan string
	disconnect chan struct{}
	channels   []string
}

func New(conf *viper.Viper) (core.Connector, error) {
	hostp, nick, ident, channels, usetls, err := ircConf(conf)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(hostp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host %q: %v", hostp, err)
	}

	tlsconf := &tls.Config{ServerName: host}

	if conf.IsSet("insecure") {
		tlsconf.InsecureSkipVerify = conf.GetBool("insecure")
	}

	cfg := &client.Config{
		Me: &state.Nick{
			Nick:  nick,
			Ident: ident,
			Name:  nick,
		},
		NewNick:     func(s string) string { return s + "_" },
		PingFreq:    3 * time.Minute,
		QuitMessage: "quitting",
		Server:      hostp,
		SSL:         usetls,
		SSLConfig:   tlsconf,
		Version:     "glenda",
		Recover:     func(conn *client.Conn, line *client.Line) {},
		SplitLen:    450,
	}

	conn := client.Client(cfg)
	conn.EnableStateTracking()

	i := &irc{
		conn:       conn,
		reconnect:  true,
		quit:       make(chan string),
		disconnect: make(chan struct{}),
		channels:   channels,
	}

	conn.HandleFunc("connected", i.connected)
	conn.HandleFunc("disconnected", i.disconnected)

	return i, nil
}

func (i *irc) String() string {
	proto := "irc"
	if i.conn.Config().SSL {
		proto = "ircs"
	}

	return fmt.Sprintf("%s://%s", proto, i.conn.Config().Server)
}

func (i *irc) running() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.reconnect
}

func (i *irc) Nick() string {
	return i.conn.Me().Nick
}

func (i *irc) Run() error {
	var lasterr error

loop:
	for i.running() {
		log.Printf("connecting to %s", i)
		if err := i.conn.Connect(); err != nil {
			lasterr = err
			log.Printf("connection error: %v", err)
			time.Sleep(10 * time.Second)
			continue

		}

		select {
		case <-i.disconnect:
			log.Printf("disconnected")
			time.Sleep(2 * time.Second)
			continue
		case msg := <-i.quit:
			i.conn.Quit(msg)
			break loop
		}
	}

	return lasterr
}

func (i *irc) Quit(msg string) {
	i.mu.Lock()
	i.reconnect = false
	i.mu.Unlock()

	if !i.conn.Connected() {
		return
	}

	i.quit <- msg
	close(i.quit)
}

func (i *irc) Message(who, msg string) {
	i.conn.Privmsg(who, msg)
}

var mmap = map[core.MessageType]string{
	core.Message:   client.PRIVMSG,
	core.Connected: client.CONNECTED,
}

func (i *irc) Handle(what core.MessageType, h core.EventHandler) {
	mtype, ok := mmap[what]
	if !ok {
		log.Printf("unhandled message type %v", what)
		return
	}

	fun := func(conn *client.Conn, line *client.Line) {
		var target string
		var e *core.Event

		if line.Args != nil && len(line.Args) > 1 {
			if strings.HasPrefix(line.Args[0], "#") {
				target = line.Args[0]
			} else {
				target = line.Nick
			}

			e = &core.Event{
				Sender: line.Nick,
				Target: target,
				Args:   strings.TrimSpace(line.Args[1]),
			}

			log.Printf("%s <%s> %s", e.Target, e.Sender, e.Args)
		}

		h.Handle(context.TODO(), i, e)
	}

	i.conn.HandleFunc(mtype, fun)
}

func (i *irc) connected(conn *client.Conn, line *client.Line) {
	for _, c := range i.channels {
		conn.Join(c)
	}
}

func (i *irc) disconnected(conn *client.Conn, line *client.Line) {
	i.disconnect <- struct{}{}
}
