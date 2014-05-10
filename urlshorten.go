package main

import (
	urlshortener "code.google.com/p/google-api-go-client/urlshortener/v1"
	"github.com/kballard/goirc/irc"
	"log"
	"net/http"
	"strings"
)

func init() {
	RegisterModule("urlshortener", func() Module {
		return &UrlShortenerMod{}
	})
}

type UrlShortenerMod struct {
	svc *urlshortener.Service
}

func (u *UrlShortenerMod) Init(b *Bot, conn irc.SafeConn) (err error) {
	u.svc, err = urlshortener.New(http.DefaultClient)
	if err != nil {
		return
	}

	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		args := strings.Split(l.Args[1], " ")
		if args[0] == ".short" {
			if len(args) < 2 {
				return
			}
			url := args[1]
			short, err := u.shorten(url)

			if err != nil {
				short = err.Error()
			}

			if l.Args[0][0] == '#' {
				c.Privmsg(l.Args[0], short)
			} else {
				c.Privmsg(l.Src.String(), short)
			}
		}
	})

	log.Printf("urlshortener module initialized")

	return nil
}

func (u *UrlShortenerMod) Reload() error {
	return nil
}

func (u *UrlShortenerMod) Call(args ...string) error {
	return nil
}

func (u *UrlShortenerMod) shorten(url string) (string, error) {
	short, err := u.svc.Url.Insert(&urlshortener.Url{
		LongUrl: url,
	}).Do()

	if err != nil {
		return "", err
	}

	return short.Id, err
}
