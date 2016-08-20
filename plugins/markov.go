package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strings"

	"github.com/mischief/glenda/markov"

	"github.com/kballard/goirc/irc"
)

func init() {
	RegisterModule("markov", func() Module {
		return &MarkovMod{}
	})
}

type MarkovMod struct {
	chain *markov.Chain
}

func (m *MarkovMod) Init(b *Bot, conn irc.SafeConn) error {
	//conf := b.Config.Search("mod", "markov")

	c, err := markov.NewChain(filepath.Join(b.DataDir, "markov"))
	if err != nil {
		return fmt.Errorf("error opening db: %s", err)
	}

	m.chain = c

	generate := func() string {
		return m.chain.Generate(rand.Intn(10) + 10)
	}

	b.Hook("markov", func(b *Bot, sender, cmd string, args ...string) error {
		b.Conn.Privmsg(sender, generate())
		return nil
	})

	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		if strings.HasPrefix(l.Args[0], b.Magic) {
			return
		}

		getAddressee := func(line string) string {
			if s := strings.SplitN(line, " ", 2); len(s) == 2 {

				t := rune(s[0][len(s[0])-1])

				if strings.ContainsRune(":,", t) {
					return s[0][:len(s[0])-1]
				}
			}

			return ""
		}

		if addressee := getAddressee(l.Args[1]); addressee != "" {
			if addressee == c.Me().String() {
				c.Privmsg(l.Args[0], generate())
			}
		} else {
			m.chain.Build(strings.NewReader(l.Args[1]))
		}
	})

	log.Printf("markov module initialized")
	return nil
}

func (m *MarkovMod) Reload() error {
	return nil
}

func (m *MarkovMod) Call(args ...string) error {
	return nil
}
