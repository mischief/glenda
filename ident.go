package main

import (
	"fmt"
	"github.com/kballard/goirc/irc"
	"log"
)

func init() {
	RegisterModule("ident", func() Module {
		return &IdentMod{}
	})
}

type IdentMod struct {
	nick, pass string
}

func (i *IdentMod) Init(b *Bot, conn irc.SafeConn) (err error) {
	conf := b.Config.Search("mod", "ident")
	i.nick = conf.Search("nick")
	if i.nick == "" {
		i.nick = b.IrcConfig.Nick
	}
	i.pass = conf.Search("pass")

	if i.pass == "" {
		log.Printf("ident: no pass")
		return nil
	}

	conn.AddHandler(irc.CONNECTED, func(c *irc.Conn, l irc.Line) {
		c.Privmsg("nickserv", fmt.Sprintf("identify %s %s", i.nick, i.pass))
	})

	return nil
}

func (i *IdentMod) Reload() error {
	return nil
}

func (i *IdentMod) Call(args ...string) error {
	return nil
}
