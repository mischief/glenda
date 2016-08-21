package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mischief/glenda/bot"

	_ "github.com/mischief/glenda/modules/define"
	_ "github.com/mischief/glenda/modules/fortune"
	_ "github.com/mischief/glenda/modules/identify"
	_ "github.com/mischief/glenda/modules/markov"
)

var (
	root = &cobra.Command{
		Use: "glenda",
		Run: runGlenda,
	}

	configfile string
)

func init() {
	root.Flags().StringVar(&configfile, "conf", "", "config file")
	rand.Seed(time.Now().UTC().UnixNano())
}

func getConf() (*viper.Viper, error) {
	conf := viper.New()
	conf.SetConfigName("glenda")

	conf.AddConfigPath("/etc/glenda/")
	conf.AddConfigPath("$HOME/.glenda")
	conf.AddConfigPath(".")

	if configfile != "" {
		conf.SetConfigFile(configfile)
	}

	err := conf.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return conf, nil
}

func run() error {
	conf, err := getConf()
	if err != nil {
		return err
	}

	bot, err := bot.New(conf)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		if err := bot.Run(); err != nil {
			log.Printf("bot error: %v", err)
		}
	}()

	<-c

	return bot.Close()
}

func runGlenda(cmd *cobra.Command, args []string) {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
