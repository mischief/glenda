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
	// available modules
	mods = make(map[string]func() Module)
	// loaded modules
	loaded = make(map[string]Module)
)

// Registers a function capable of creating a new Module instance.
func RegisterModule(name string, f func() Module) {
	mods[name] = f
}

func LoadModule(name string) Module {
	if m, ok := mods[name]; ok {
		loaded[name] = m()
		return loaded[name]
	}

	return nil
}

func GetModule(name string) Module {
	if m, ok := loaded[name]; ok {
		return m
	}
	return nil
}

func IsLoaded(name string) bool {
	if _, ok := loaded[name]; ok {
		return true
	}

	return false
}

func ListModules() []string {
	var out []string
	for n, _ := range mods {
		out = append(out, n)
	}

	return out
}

func init() {
	RegisterModule("module", func() Module {
		return &ModMod{}
	})
}

type ModMod struct {
}

func (m *ModMod) Init(b *Bot, conn irc.SafeConn) error {
	return nil
}

func (m *ModMod) Reload() error {
	return nil
}

func (m *ModMod) Call(args ...string) error {
	return nil
}
