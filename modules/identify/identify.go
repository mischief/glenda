// +build irc

package identify

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/mischief/glenda/core"
	"github.com/mischief/glenda/util"
)

var mod = &core.ModInfo{
	Name:   "identify",
	Create: create,
}

func init() {
	core.RegisterModule(mod)
}

func create(b core.Bot, conf *viper.Viper) (core.Module, error) {
	if err := util.RequiredOptions(conf, "nick", "pass"); err != nil {
		return nil, err
	}

	m := &identify{
		nick: conf.GetString("nick"),
		pass: conf.GetString("pass"),
	}

	return m, nil
}

type identify struct {
	nick string
	pass string
}

func (i *identify) Close() error { return nil }
func (i *identify) Name() string { return mod.Name }

func (i *identify) GetHandlers() map[core.EventType]core.EventHandler {
	m := map[core.EventType]core.EventHandler{}

	et := core.EventMessage{core.Connected}

	m[et] = i

	return m
}

func (i *identify) Handle(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	log.Printf("identifying as %s", i.nick)
	rw.Message("NickServ", fmt.Sprintf("IDENTIFY %s %s", i.nick, i.pass))
	return nil
}
