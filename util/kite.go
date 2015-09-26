package util

import (
	"net/url"

	"github.com/koding/kite"
	"github.com/koding/kite/config"
)

func NewKite(name, version string) *kite.Kite {
	var self *url.URL

	k := kite.New(name, version)
	k.Config = config.MustGet()

	go k.Run()
	<-k.ServerReadyNotify()
	k.Config.Port = k.Port()

	self = k.RegisterURL(true)
	self.Path = "/kite"

	k.RegisterForever(self)
	return k
}

func ArgSlice(r *kite.Request) []string {
	var args []string
	slice := r.Args.MustSlice()[0].MustSlice()

	for _, p := range slice {
		s := p.MustString()
		args = append(args, s)
	}

	return args
}
