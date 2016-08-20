package main

import (
	"encoding/json"
	"fmt"
	"github.com/kballard/goirc/irc"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	RegisterModule("define", func() Module {
		return &DefineMod{}
	})
}

type UrbanWord struct {
	List []struct {
		Author      string  `json:"author"`
		CurrentVote string  `json:"current_vote"`
		Defid       float64 `json:"defid"`
		Definition  string  `json:"definition"`
		Example     string  `json:"example"`
		Permalink   string  `json:"permalink"`
		ThumbsDown  float64 `json:"thumbs_down"`
		ThumbsUp    float64 `json:"thumbs_up"`
		Word        string  `json:"word"`
	} `json:"list"`
	ResultType string        `json:"result_type"`
	Sounds     []interface{} `json:"sounds"`
	Tags       []string      `json:"tags"`
}

type DefineMod struct {
}

func (d *DefineMod) Init(b *Bot, conn irc.SafeConn) error {

	b.Hook("define", func(b *Bot, sender, cmd string, args ...string) error {

		var (
			i   = 0
			err error
		)

		if len(args) > 1 {
			if i, err = strconv.Atoi(args[len(args)-1]); err == nil {
				args = args[:len(args)-1]
			}
		}

		b.Conn.Privmsg(sender, d.define(strings.Join(args, "+"), i))
		return nil
	})

	log.Printf("define module initialized with cmd define")
	return nil
}

func (d *DefineMod) define(word string, i int) string {
	var definition UrbanWord
	var body []byte

	url := fmt.Sprintf("http://api.urbandictionary.com/v0/define?term=%s", word)

	resp, err := http.Get(url)
	if err != nil {
		goto bad
	}

	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		goto bad
	}

	if err = json.Unmarshal(body, &definition); err != nil {
		goto bad
	}

	if len(definition.List) > 0 {

		if i >= len(definition.List) || i < 0 {
			i = 0
		}

		return fmt.Sprintf("%s [%d/%d]: %s", word, i, len(definition.List)-1, strings.Replace(definition.List[i].Definition, "\r\n", " ", -1))
	} else {
		return fmt.Sprintf("%s: not found.", word)
	}

bad:
	return fmt.Sprintf("define error: %s", err)
}

func (m *DefineMod) Reload() error {
	return nil
}

func (m *DefineMod) Call(args ...string) error {
	return nil
}
