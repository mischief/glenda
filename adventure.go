
package main

import (
	"bufio"
	"github.com/kballard/goirc/irc"
	"io"
	"log"
	"os/exec"
	"strings"
  "time"
  "fmt"
)

func init() {
	RegisterModule("adventure", func() Module {
		return &AdventureMod{}
	})
}

type AdventureMod struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Scanner //io.ReadCloser
}

func (g *AdventureMod) Init(b *Bot, conn irc.SafeConn) error {
	conf := b.Config.Search("mod", "adventure")
	channel := conf.Search("channel")

	if err := g.spawn(); err != nil {
		return err
	}

	go func() {
    time.Sleep(10 * time.Second)

		for g.out.Scan() {
			line := g.out.Text()
			log.Printf("adventure: %s", line)
			conn.Privmsg(channel, line)
		}

		if err := g.out.Err(); err != nil {
			log.Printf("adventure read error: %s", err)
		}
	}()

	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		args := strings.Split(l.Args[1], " ")
		if args[0] == ".a" {
			line := strings.Join(args[1:], " ")
			log.Printf("adventure: writing %q", line)
			if _, err := fmt.Fprintf(g.in, "%s\n", line); err != nil {
				log.Printf("adventure: error writing to subprocess: %s", err)
			}
		}
	})

	log.Printf("adventure module initialized with channel %s", channel)

	return nil
}

func (g *AdventureMod) Reload() error {
	return nil
}

func (g *AdventureMod) Call(args ...string) error {
	return nil
}

func (g *AdventureMod) spawn() error {
	g.cmd = exec.Command("unbuffer", "-p", "adventure")

	out, err := g.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	g.out = bufio.NewScanner(out)

	g.in, err = g.cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err = g.cmd.Start(); err != nil {
		return err
	}

	// get rid of prompt
	io.WriteString(g.in, "n\n")

  g.out.Scan()
  g.out.Scan()
  g.out.Scan()

	return nil
}
