package main

import (
	oauth "golang.org/x/oauth2"
	urlshortener "google.golang.org/api/urlshortener/v1"
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
			Endpoint: oauth.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://accounts.google.com/o/oauth2/token",
			},
		}
		return &GoogleApiMod{config: oc}
	})
}

type GoogleApiMod struct {
	config *oauth.Config
}

func (g *GoogleApiMod) Init(b *Bot, conn irc.SafeConn) error {
	conf := b.Config.Search("mod", "googleapi")
	g.config.ClientID = conf.Search("clientid")
	g.config.ClientSecret = conf.Search("clientsecret")
	g.config.Scopes = []string{urlshortener.UrlshortenerScope}

	log.Printf("googleapi module initialized")
	return nil
}

func (g *GoogleApiMod) Reload() error {
	return nil
}

func (g *GoogleApiMod) Call(args ...string) error {
	return nil
}
