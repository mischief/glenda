package main

import (
	"fmt"
	"github.com/kballard/goirc/irc"
	urlshortener "google.golang.org/api/urlshortener/v1"
	"log"
	"net/http"
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

	b.Hook("short", func(b *Bot, sender, cmd string, args ...string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing argument")
		}
		short, err := u.shorten(args[0])
		if err != nil {
			return err
		}

		b.Conn.Privmsg(sender, short)
		return nil
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
