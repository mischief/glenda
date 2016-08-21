package core

import (
	"testing"

	"github.com/spf13/viper"
)

func TestModuleRegistryGet(t *testing.T) {
	testModule := &ModInfo{
		Name: "test",
	}

	r := newRegistry()

	m, err := r.Get("test")
	if err == nil {
		t.Fatalf("got module, expected none: %+v", m)
	}

	r.Register(testModule)

	m, err = r.Get("test")
	if err != nil {
		t.Fatalf("didn't find module %q: %v", err)
	}
}

type testModule struct {
	b             Bot
	conf          *viper.Viper
	closeerr      error
	name          string
	eventhandlers map[EventType]EventHandler
}

func (t *testModule) Close() error { return t.closeerr }
func (t *testModule) Name() string { return t.name }
func (t *testModule) GetHandlers() map[EventType]EventHandler {
	return t.eventhandlers
}

func testModuleCreate(b Bot, conf *viper.Viper) (Module, error) {
	return &testModule{b: b, conf: conf}, nil
}

func TestModuleRegistryCreate(t *testing.T) {
	n := "test"
	r := newRegistry()
	mi, m, err := r.Create(n, nil, nil)
	if err == nil {
		t.Fatalf("created module, expected none: %+v", mi)
	}

	r.Register(&ModInfo{n, testModuleCreate})

	mi, m, err = r.Create(n, nil, nil)
	if err != nil {
		t.Fatalf("failed to create module %q: %v", n, err)
	}

	if err := m.Close(); err != nil {
		t.Fatalf("failed to close module %q: %v", n, err)
	}
}
