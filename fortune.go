package main

import (
	"fmt"
	"github.com/kballard/goirc/irc"
	"log"
	"os/exec"
	"strings"
)

func init() {
	RegisterModule("fortune", func() Module {
		return &FortuneMod{}
	})
}

type FortuneMod struct {
	cmd []string
}

func (f *FortuneMod) Init(b *Bot, conn irc.SafeConn) error {
	f.cmd = []string{"fortune", "-e", "-a", "-s"}

	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		if l.Args[1] == ".fortune" {
			str := strings.Replace(f.fortune(), "\t", " ", -1)
			strs := strings.Split(str, "\n")

			log.Printf("fortune %+v", strs)
			for _, s := range strs {
				if l.Args[0][0] == '#' {
					c.Privmsg(l.Args[0], s)
				} else {
					c.Privmsg(l.Src.String(), s)
				}
			}
		}
	})

	log.Printf("forutne module initialized with cmd %s", strings.Join(f.cmd, " "))
	return nil
}

func (f *FortuneMod) fortune() string {
	c := exec.Command(f.cmd[0], f.cmd[1:]...)

	out, err := c.CombinedOutput()

	if err != nil {
		return fmt.Sprintf("fortune error: %S", err)
	} else {
		return string(out)
	}
}

func (m *FortuneMod) Reload() error {
	return nil
}

func (m *FortuneMod) Call(args ...string) error {
	return nil
}
