package main

import (
	"fmt"
	"github.com/kballard/goirc/irc"
	"strings"
	"time"
)

func init() {
	RegisterModule("time", func() Module {
		return &TimeMod{}
	})
}

type TimeMod struct {
}

func (t *TimeMod) Init(b *Bot, conn irc.SafeConn) (err error) {
	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		args := strings.Split(l.Args[1], " ")
		if args[0] == ".time" {
			t := time.Now()
			if len(args) == 2 {
				tz := args[1]
				loc, err := time.LoadLocation(tz)
				if err != nil {
					return
				}
				t = t.In(loc)
			}
			if l.Args[0][0] == '#' {
				c.Privmsg(l.Args[0], fmt.Sprintf("%s", t))
			} else {
				c.Privmsg(l.Src.String(), fmt.Sprintf("%s", t))
			}
		}
	})

	return nil
}

func (t *TimeMod) Reload() error {
	return nil
}

func (t *TimeMod) Call(args ...string) error {
	return nil
}
