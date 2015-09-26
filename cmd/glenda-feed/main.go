package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/mischief/glenda/util"

	//"github.com/SlyMarbo/rss"
	//"github.com/koding/kite"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type RSSConfig struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Channels []string `json:"channels"`
	Color    string   `json:"color"`
}

type Config struct {
	Feeds []RSSConfig `json:"feeds"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	etcd = flag.String("etcd-endpoint", "http://127.0.0.1:2379", "etcd endpoint url")
	conf = flag.String("config", "/offblast/feed/config.json", "configuration key in etcd")
)

func main() {
	var config Config

	flag.Parse()
	k := util.NewKite("feed", "1.0.0")

	viper := viper.New()
	viper.AddRemoteProvider("etcd", *etcd, *conf)
	viper.SetConfigType("json")

	if err := viper.ReadRemoteConfig(); err != nil {
		k.Log.Fatal("Failed to retrieve config: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		k.Log.Fatal("Failed to unmarshal config: %v", err)
	}

	// TODO(mischief): port feed code

	irc, err := util.IrcDial(k)
	if err != nil {
		k.Log.Fatal("Failed to dial irc: %v", err)
	}

	out := map[string]string{
		"reply-to": "#feed",
		"message":  util.IrcColor("Feed reader initialized!", "lime"),
	}

	irc.Tell("irc.privmsg", out)

	//k.HandleFunc("theo", theo).DisableAuthentication()
	<-k.ServerCloseNotify()
}
