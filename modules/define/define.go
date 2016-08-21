// Package define gets definitions of words or phrases from
// urbandictionary.com.

package define

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/mischief/glenda/core"
)

func init() {
	mod := &core.ModInfo{
		Name: "define",
	}

	mod.Create = func(b core.Bot, conf *viper.Viper) (core.Module, error) {
		return &core.ModuleFunc{mod.Name, urbandictionary}, nil
	}

	core.RegisterModule(mod)
}

func urbandictionary(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var err error
	args := strings.Split(e.Args, " ")
	i := 1

	if len(args) < 1 {
		return nil
	}

	if len(args) > 1 {
		if i, err = strconv.Atoi(args[len(args)-1]); err == nil {
			args = args[:len(args)-1]
		}
	}

	var msg string
	def, err := define(ctx, strings.Join(args, " "), i)
	if err != nil {
		msg = fmt.Sprintf("error: %v", err)
	} else {
		msg = def
	}

	rw.Message(e.Target, msg)
	return nil
}

type urbanWord struct {
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

func define(ctx context.Context, word string, i int) (string, error) {
	escaped := url.QueryEscape(word)
	url := fmt.Sprintf("http://api.urbandictionary.com/v0/define?term=%s", escaped)

	resp, err := ctxhttp.Get(ctx, http.DefaultClient, url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var definition urbanWord

	if err := json.NewDecoder(resp.Body).Decode(&definition); err != nil {
		return "", err
	}

	defs := definition.List

	if len(defs) < 1 {
		return "", fmt.Errorf("no definition found for %q")
	}

	if i > len(defs) || i < 1 {
		i = 1
	}

	def := defs[i-1].Definition
	san := strings.Replace(def, "\r\n", " ", -1)

	return fmt.Sprintf("%s [%d/%d]: %s", word, i, len(defs), san), nil
}
