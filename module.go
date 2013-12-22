package main

import (
	"github.com/kballard/goirc/irc"
)

type Module interface {
	Init(*Bot, irc.SafeConn) error
	Reload() error
	Call(args ...string) error
}

var (
	mods = make(map[string]func() Module)
)

// Registers a function capable of creating a new Module instance.
func RegisterModule(name string, f func() Module) {
	mods[name] = f
}

func LoadModule(name string) Module {
	if m, ok := mods[name]; ok {
		return m()
	}

	return nil
}

func ListModules() []string {
	var out []string
	for n, _ := range mods {
		out = append(out, n)
	}

	return out
}
