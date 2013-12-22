package main

import (
	"fmt"
	"github.com/kballard/goirc/irc"
	"github.com/luksen/maildir"
	"github.com/robfig/cron"
	"log"
	"time"
)

func init() {
	RegisterModule("mailwatch", func() Module {
		return &MailwatchMod{}
	})
}

type MailwatchMod struct {
	dir     maildir.Dir
	cronjob *cron.Cron
}

func (m *MailwatchMod) Init(b *Bot, conn irc.SafeConn) error {
	conf := b.Config.Search("mod", "mailwatch")
	dir := conf.Search("dir")
	channel := conf.Search("channel")

	if dir != "" {
		m.dir = maildir.Dir(dir)
		m.cronjob = cron.New()

		m.cronjob.AddFunc("@every 1m", func() {
			//log.Printf("checking mail %s...", m.dir)

			if newmail, err := m.dir.Unseen(); err != nil {
				conn.Privmsg(channel, fmt.Sprintf("maildir error: %s, err"))
			} else {
				l := len(newmail)

				if l > 0 {
					conn.Privmsg(channel, fmt.Sprintf("%d new mail:", l))

					for _, k := range newmail {
						hdr, err := m.dir.Header(k)
						if err != nil {
							conn.Privmsg(channel, fmt.Sprintf("maildir header error: %s", err))
						} else {
							conn.Privmsg(channel, fmt.Sprintf("from   : %s", hdr.Get("From")))
							conn.Privmsg(channel, fmt.Sprintf("subject: %s", hdr.Get("Subject")))
						}
					}
				}
			}
		})

		go func() {
			time.Sleep(10 * time.Second)
			m.cronjob.Start()
		}()

		log.Printf("mailwatch module initialized with dir: %s", dir)
	}

	return nil
}

func (m *MailwatchMod) Reload() error {
	return nil
}

func (m *MailwatchMod) Call(args ...string) error {
	return nil
}
