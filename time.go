package main

import (
	"fmt"
	"time"

	"github.com/kballard/goirc/irc"
)

func init() {
	RegisterModule("time", func() Module {
		return &TimeMod{}
	})
}

type TimeMod struct {
}

func (t *TimeMod) Init(b *Bot, conn irc.SafeConn) (err error) {
	b.Hook("time", func(b *Bot, sender, cmd string, args ...string) error {
		t := time.Now()
		if len(args) == 1 {
			tz := args[0]
			loc, err := time.LoadLocation(tz)
			if err != nil {
				return err
			}
			t = t.In(loc)
		}

		b.Conn.Privmsg(sender, fmt.Sprintf("%s", t))
		return nil
	})

	return nil
}

func (t *TimeMod) Reload() error {
	return nil
}

func (t *TimeMod) Call(args ...string) error {
	return nil
}
