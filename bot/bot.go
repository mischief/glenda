package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/mischief/glenda/connector"
	"github.com/mischief/glenda/core"

	// bind in connectors
	_ "github.com/mischief/glenda/connector/irc"
)

type Bot struct {
	conf *viper.Viper
	c    core.Connector

	// command trigger
	magic string

	// data directory
	datadir string

	mmu     sync.Mutex
	modules map[string]core.Module
}

func getConnector(conf *viper.Viper) (core.Connector, error) {
	if !conf.IsSet("core.connect") {
		return nil, fmt.Errorf("core.connector not set!")
	}

	cname := conf.GetString("core.connect")

	if !conf.IsSet("network." + cname) {
		return nil, fmt.Errorf("network %q is not configured", cname)
	}

	if !conf.IsSet("network." + cname + ".type") {
		return nil, fmt.Errorf("network %q has no type", cname)
	}

	ctype := conf.GetString("network." + cname + ".type")

	cconf := conf.Sub("network." + cname)

	return connector.Create(ctype, cconf)
}

func New(conf *viper.Viper) (*Bot, error) {
	con, err := getConnector(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connector: %v", err)
	}

	if !conf.IsSet("core.magic") {
		return nil, fmt.Errorf("missing core.magic!")
	}

	datadir := "."

	if conf.IsSet("core.datadir") {
		datadir = conf.GetString("core.datadir")
	}

	bot := &Bot{
		conf:    conf,
		c:       con,
		magic:   conf.GetString("core.magic"),
		datadir: datadir,
		modules: make(map[string]core.Module),
	}

	if err := bot.initmodules(); err != nil {
		bot.Close()
		return nil, err
	}

	con.Handle(core.Message, core.EventHandlerFunc(bot.msghandler))
	con.Handle(core.Connected, core.EventHandlerFunc(bot.connectedhandler))

	return bot, nil
}

func (b *Bot) Close() error {
	log.Printf("closing modules")

	b.mmu.Lock()
	for _, m := range b.modules {
		m.Close()
	}

	b.mmu.Unlock()

	log.Printf("closing connector")

	b.c.Quit("exiting")

	return nil
}

func (b *Bot) initmodules() error {
	// init modules
	if !b.conf.IsSet("core.modules") {
		return nil
	}

	modnames := b.conf.GetStringSlice("core.modules")

	b.mmu.Lock()
	defer b.mmu.Unlock()

	for _, name := range modnames {
		var mconf *viper.Viper
		if b.conf.IsSet("modules." + name) {
			mconf = b.conf.Sub("modules." + name)
		} else {
			mconf = viper.New()
		}

		log.Printf("loading module %q", name)

		_, m, err := core.CreateModule(name, b, mconf)
		if err != nil {
			return err
		}

		b.modules[name] = m
	}

	return nil
}

func (b *Bot) Run() error {
	log.Printf("running with network %s", b.c)
	return b.c.Run()
}

func (b *Bot) iscmd(e *core.Event) (bool, string, string) {
	args := strings.Split(e.Args, " ")
	cmd := args[0]

	if !strings.HasPrefix(cmd, b.magic) {
		return false, "", ""
	}

	cmd = strings.TrimPrefix(cmd, b.magic)
	newargs := strings.Join(args[1:], " ")

	return true, cmd, newargs
}

// Event handlers for events from connectors
func (b *Bot) msghandler(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	iscmd, cmd, args := b.iscmd(e)

	for _, mod := range b.modules {
		for typ, h := range mod.GetHandlers() {
			switch et := typ.(type) {
			case core.EventMessage:
				if et.Type == core.Message {
					h.Handle(ctx, rw, e)
				}
			case core.EventCommand:
				if iscmd && et.Command == cmd {
					newe := *e
					newe.Args = args
					h.Handle(ctx, rw, &newe)
				}
			}
		}
	}

	return nil
}

func (b *Bot) connectedhandler(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	log.Printf("connected")
	for _, mod := range b.modules {
		for typ, h := range mod.GetHandlers() {
			if em, ok := typ.(core.EventMessage); ok {
				if em.Type == core.Connected {
					h.Handle(ctx, rw, e)
				}
			}
		}
	}

	return nil
}

func (b *Bot) Open(m core.Module, name string) (core.File, error) {
	p := filepath.Clean(name)

	return os.Open(filepath.Join(b.datadir, m.Name(), p))
}

func (b *Bot) Path(m core.Module) string {
	return filepath.Join(b.datadir, m.Name())
}

func (b *Bot) Magic() string { return b.magic }
func (b *Bot) Nick() string  { return b.c.Nick() }
