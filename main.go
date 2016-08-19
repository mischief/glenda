package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/time/rate"
	"github.com/kballard/goirc/irc"
	"github.com/mischief/ndb"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type HookFn func(b *Bot, sender, cmd string, args ...string) error

type Bot struct {
	Channels  []string
	Config    *ndb.Ndb
	Conn      irc.SafeConn
	IrcConfig irc.Config
	Mods      map[string]Module
	Magic     string
	DataDir   string

	LoginFn   func(conn *irc.Conn, line irc.Line)
	PrivmsgFn func(conn *irc.Conn, line irc.Line)
	ActionFn  func(conn *irc.Conn, line irc.Line)

	ratelimit map[string]*rate.Limiter

	hooks map[string]HookFn

	quit chan bool
}

func NewBot(conf string) (*Bot, error) {
	var err error

	bot := &Bot{
		Mods: make(map[string]Module),
		quit: make(chan bool, 1),
	}

	bot.LoginFn = func(conn *irc.Conn, line irc.Line) {
		for _, c := range bot.Channels {
			conn.Join([]string{c}, nil)
		}
	}

	bot.PrivmsgFn = func(conn *irc.Conn, l irc.Line) {
		log.Printf("[%s] %s> %s\n", l.Args[0], l.Src, l.Args[1])
		/*
			if l.Args[1] == ".quit" {
				conn.Quit("quit")
				return
			}
		*/

		args := strings.Split(l.Args[1], " ")
		cmd := args[0]

		args = args[1:]

		if cmd == "" {
			return
		}

		if !strings.HasPrefix(cmd, bot.Magic) {
			return
		}

		cmd = strings.TrimPrefix(cmd, bot.Magic)

		hk, ok := bot.hooks[cmd]
		if !ok {
			return
		}

		var sender string

		if l.Args[0][0] == '#' {
			sender = l.Args[0]
		} else {
			sender = l.Src.String()
		}

		if limiter, ok := bot.ratelimit[l.Args[0]]; ok {
			if !limiter.Allow() {
				log.Printf(`rate limiting`)
				return
			}
		}

		if err := hk(bot, sender, cmd, args...); err != nil {
			bot.Conn.Privmsg(sender, err.Error())
		}
	}

	bot.ActionFn = func(conn *irc.Conn, line irc.Line) {
		log.Printf("[%s] %s %s\n", line.Dst, line.Src, line.Args[0])
	}

	bot.hooks = make(map[string]HookFn)

	config, err := ndb.Open(conf)
	if err != nil {
		log.Fatalf("cannot open config file %s: %s", *configfile, err)
	}

	bot.Config = config

	ircconf := config.Search("irc", "")

	if len(ircconf) <= 0 {
		log.Fatalf("missing irc section in config %s", *configfile)
	}

	bot.IrcConfig, err = bot.parseconfig(ircconf)
	if err != nil {
		log.Fatalf("error parsing config: %s", err)
	}

	//log.Printf("bot config: %+v", bot.IrcConfig)
	log.Printf("channels: %+v", bot.Channels)

	log.Printf("modules available: %s", strings.Join(ListModules(), " "))

	var mods []string
	for n, _ := range bot.Mods {
		mods = append(mods, n)
	}

	log.Printf("modules loaded: %s", strings.Join(mods, " "))

	bot.IrcConfig.Init = func(hr irc.HandlerRegistry) {
		log.Println("initializing...")
		hr.AddHandler(irc.CONNECTED, bot.LoginFn)
		hr.AddHandler(irc.DISCONNECTED, func(*irc.Conn, irc.Line) {
			fmt.Println("disconnected")
			bot.quit <- true
		})
		hr.AddHandler("PRIVMSG", bot.PrivmsgFn)
		hr.AddHandler(irc.ACTION, bot.ActionFn)
	}

	return bot, err
}

func (b *Bot) Run() (err error) {
	log.Println("connecting...")
	if b.Conn, err = irc.Connect(b.IrcConfig); err != nil {
		close(b.quit)
		return err
	}

	for _, m := range b.Mods {
		if err = m.Init(b, b.Conn); err != nil {
			return
		}
	}

	<-b.quit
	log.Println("goodbye.")

	return
}

func (b *Bot) Conf() *ndb.Ndb {
	return b.Config
}

// Reload config and reconfigure bot
func (b *Bot) Reload() error {
	if err := b.Config.Reopen(); err != nil {
		return err
	}

	for n, m := range b.Mods {
		if err := m.Reload(); err != nil {
			log.Printf("module %s failed to reload: %s", n, err)
		}
	}

	return nil
}

func (b *Bot) parseconfig(c ndb.RecordSet) (irc.Config, error) {
	var err error
	var limits, bursts string
	var limit float64
	var burst int64

	hosts := c.Search("host")
	ports := c.Search("port")
	ssls := c.Search("ssl")
	nicks := c.Search("nick")
	users := c.Search("user")
	reals := c.Search("real")
	floods := c.Search("flood")
	channelss := c.Search("channels")
	moduless := c.Search("modules")
	magics := c.Search("magic")
	datadirs := c.Search("datadir")

	conf := irc.Config{
		Host:      hosts,
		Nick:      nicks,
		User:      users,
		RealName:  reals,
		SSLConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if port, err := strconv.Atoi(ports); err != nil {
		goto badconf
	} else {
		conf.Port = uint(port)
	}

	if floods == "true" {
		conf.AllowFlood = true
	} else {
		conf.AllowFlood = true
	}

	if ssls == "true" {
		conf.SSL = true
	} else {
		conf.SSL = false
	}

	b.Channels = strings.Split(channelss, " ")

	b.ratelimit = make(map[string]*rate.Limiter)
	limits = c.Search("ratelimit_rate")
	if limits == "" {
		limits = "1"
	}

	bursts = c.Search("ratelimit_burst")
	if bursts == "" {
		bursts = "1"
	}

	limit, err = strconv.ParseFloat(limits, 64)
	if err != nil {
		goto badconf
	}
	burst, err = strconv.ParseInt(bursts, 10, 64)
	if err != nil {
		goto badconf
	}
	for _, c := range b.Channels {
		b.ratelimit[c] = rate.NewLimiter(rate.Limit(limit), int(burst))
	}

	if mods := strings.Split(moduless, " "); len(mods) > 0 {
		for _, m := range mods {
			if mod := LoadModule(m); mod != nil {
				b.Mods[m] = mod
			} else {
				log.Printf("no such module %s", m)
			}
		}
	}

	if magics != "" {
		b.Magic = magics
	} else {
		b.Magic = "."
	}

	if datadirs != "" {
		b.DataDir = datadirs
	} else {
		b.DataDir = os.ExpandEnv("${HOME}/.glenda")
	}

	return conf, nil
badconf:

	return conf, fmt.Errorf("config error: %s", err)
}

func (b *Bot) Hook(cmd string, hook HookFn) error {
	if _, ok := b.hooks[cmd]; ok {
		return fmt.Errorf("hook for %q already exists", cmd)
	}

	b.hooks[cmd] = hook
	return nil
}

var (
	configfile = flag.String("conf", "config/main", "path to ndb(6)-format config file")
)

func main() {
	var err error

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	bot, err := NewBot(*configfile)

	if err != nil {
		log.Fatal(err)
	}

	if bot != nil {
		log.Fatal(bot.Run())
	}
}
