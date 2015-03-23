package main

import (
	"fmt"
	"strings"
	"time"
	"github.com/kballard/goirc/irc"
)

func init () {
	RegisterModule("notify", func() Module{
		return &NotifyMod{}
	})
}

type Note struct {
	from string
	message string
	sent time.Time
}

func (n Note) String () string {
	return fmt.Sprintf("%s <%s> %s", n.sent.Format("01/02 15:04"), n.from, n.message)
}

type NotifyMod struct {
	notes map[string][]Note
}

func (m *NotifyMod) NotifyIfQueued (conn *irc.Conn, line irc.Line) {

	to := line.Src.String()

	if notes, ok := m.notes[to]; ok {
		
		ctx := getContext(line)

		for _, note := range notes {
			conn.Privmsg(ctx, fmt.Sprintf("%s: %s", to, note))
		}

		delete(m.notes, to)
	}
}

func (m *NotifyMod) Init (b *Bot, conn irc.SafeConn) error {

	m.notes = make(map[string][]Note)

	conn.AddHandler("PRIVMSG", func (con *irc.Conn, line irc.Line) {
		args := strings.SplitN(line.Args[1], " ", 3)

		if len(args) == 3 && args[0] == ".notify" {

			note := Note{
				from: line.Src.String(),
				message: args[2],
				sent: time.Now(),
			}

			m.notes[args[1]] = append(m.notes[args[1]], note);

			conn.Privmsg(getContext(line), fmt.Sprintf("added note to %q: %q", args[1], note))
		}
	})

	notify := func (conn *irc.Conn, line irc.Line) {
		m.NotifyIfQueued(conn, line)
	}

	conn.AddHandler("PRIVMSG", notify)
	conn.AddHandler("JOIN", notify)

	return nil
}

func (m *NotifyMod) Reload () error {
	return nil
}

func (m *NotifyMod) Call (args ...string) error {
	return nil
}

//-- utility

func getContext (l irc.Line) string {
	if l.Args[0][0] == '#'{
		return l.Args[0]
	}
	return l.Src.String()
}

