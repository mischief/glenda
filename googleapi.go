package main

import (
	"code.google.com/p/goauth2/oauth"
	urlshortener "code.google.com/p/google-api-go-client/urlshortener/v1"
	//"encoding/json"
	//"fmt"
	"github.com/kballard/goirc/irc"
	//"io/ioutil"
	"log"
	//"net/http"
	//"strings"
)

func init() {
	RegisterModule("googleapi", func() Module {
		oc := &oauth.Config{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		}
		return &GoogleApiMod{config: oc}
	})
}

type GoogleApiMod struct {
	config *oauth.Config
}

func (g *GoogleApiMod) Init(b *Bot, conn irc.SafeConn) error {
	conf := b.Config.Search("mod", "googleapi")
	g.config.ClientId = conf.Search("clientid")
	g.config.ClientSecret = conf.Search("clientsecret")
	g.config.Scope = urlshortener.UrlshortenerScope

	/*
		conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
			args := strings.Split(l.Args[1], " ")
			if args[0] == ".geo" {
				ip := strings.Join(args[1:], "")
				loc := g.geo(ip)

				if l.Args[0][0] == '#' {
					c.Privmsg(l.Args[0], loc)
				} else {
					c.Privmsg(l.Src.String(), loc)
				}
			}
		})
	*/

	log.Printf("googleapi module initialized")

	return nil
}

func (g *GoogleApiMod) Reload() error {
	return nil
}

func (g *GoogleApiMod) Call(args ...string) error {
	return nil
}
