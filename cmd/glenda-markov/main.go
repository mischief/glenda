package main

import (
	"github.com/mischief/glenda/util"

	"github.com/koding/kite"
)

func main() {
	k := util.NewKite("markov", "1.0.0")
	k.HandleFunc("markov.markov", markov).DisableAuthentication()
	<-k.ServerCloseNotify()
}

func markov(r *kite.Request) (result interface{}, err error) {
	return "wubalubadubdub!", nil
}
