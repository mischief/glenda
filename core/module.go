package core

import (
	"fmt"
	"io"
	"sync"

	"github.com/spf13/viper"
)

type Module interface {
	io.Closer
	Name() string
	GetHandlers() map[EventType]EventHandler
}

type ModuleCreator func(b Bot, conf *viper.Viper) (Module, error)

type ModInfo struct {
	Name   string
	Create ModuleCreator
}

type modRegistry struct {
	mu sync.Mutex
	// module factories
	mods map[string]*ModInfo
}

func newRegistry() *modRegistry {
	return &modRegistry{
		mods: make(map[string]*ModInfo),
	}
}

func (m *modRegistry) Register(mi *ModInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := mi.Name

	if _, ok := m.mods[name]; ok {
		panic("module " + name + " already exists")
	}

	m.mods[name] = mi
}

func (m *modRegistry) Get(name string) (*ModInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mi, ok := m.mods[name]
	if !ok {
		return nil, fmt.Errorf("module %q is not registered")
	}

	return mi, nil
}

func (m *modRegistry) Create(name string, b Bot, conf *viper.Viper) (*ModInfo, Module, error) {
	mi, err := m.Get(name)
	if err != nil {
		return nil, nil, err
	}

	mod, err := mi.Create(b, conf)
	if err != nil {
		return nil, nil, err
	}

	return mi, mod, nil
}

var registry = newRegistry()

func RegisterModule(m *ModInfo) {
	registry.Register(m)
}

func GetModule(name string) (*ModInfo, error) {
	return registry.Get(name)
}

func CreateModule(name string, b Bot, conf *viper.Viper) (*ModInfo, Module, error) {
	return registry.Create(name, b, conf)
}

// ModuleFunc is a convenience wrapper for modules which only have one
// function, and it is called by one name.
type ModuleFunc struct {
	// The name of the module/function
	FName string
	// Function to invoke
	F EventHandlerFunc
}

func (mf *ModuleFunc) Close() error { return nil }
func (mf *ModuleFunc) Name() string { return mf.FName }
func (mf *ModuleFunc) GetHandlers() map[EventType]EventHandler {
	return map[EventType]EventHandler{EventCommand{mf.FName}: mf.F}
}
