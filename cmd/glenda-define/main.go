package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mischief/glenda/util"

	"github.com/koding/kite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	k := util.NewKite("define", "1.0.0")
	k.HandleFunc("define.define", defineHandler).DisableAuthentication()
	<-k.ServerCloseNotify()
}

func defineHandler(r *kite.Request) (result interface{}, err error) {
	args := util.ArgSlice(r)
	i := 0

	if len(args) > 1 {
		if i, err = strconv.Atoi(args[len(args)-1]); err == nil {
			args = args[:len(args)-1]
		}
	}

	if len(args) < 1 {
		return nil, fmt.Errorf("no query given")
	}

	return define(strings.Join(args, " "), i), nil
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

func define(word string, i int) string {
	var definition UrbanWord
	var body []byte

	url := fmt.Sprintf("http://api.urbandictionary.com/v0/define?term=%s", url.QueryEscape(word))

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
