package connector

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/mischief/glenda/core"
)

type Creator func(*viper.Viper) (core.Connector, error)

var connectors = map[string]Creator{}

func Register(name string, c Creator) {
	_, ok := connectors[name]
	if ok {
		panic("creator " + name + " already exists")
	}

	connectors[name] = c
}

func Create(name string, conf *viper.Viper) (core.Connector, error) {
	c, ok := connectors[name]
	if !ok {
		return nil, fmt.Errorf("no such connector %q", name)
	}

	return c(conf)
}
