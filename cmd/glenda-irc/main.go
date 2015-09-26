package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/mischief/glenda/bot"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var (
	etcd    = flag.String("etcd-endpoint", "http://127.0.0.1:2379", "etcd endpoint url")
	prefix  = flag.String("prefix", "/offblast/glenda/modules/irc/", "configuration prefix in etcd")
	network = flag.String("network", "default", "configuration file base name in etcd")
)

func main() {
	flag.Parse()

	b, err := bot.NewBot(*etcd, fmt.Sprintf("%s/%s.json", *prefix, *network))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	b.Run()
}
