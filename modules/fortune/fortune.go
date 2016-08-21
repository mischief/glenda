package fortune

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/mischief/glenda/core"
)

var mod = &core.ModInfo{
	Name:   "fortune",
	Create: create,
}

func init() {
	core.RegisterModule(mod)
}

func create(b core.Bot, conf *viper.Viper) (core.Module, error) {
	m := &fortune{
		fortunes: make(map[string][]string),
	}

	if !conf.IsSet("files") {
		return nil, fmt.Errorf("no fortune files found")
	}

	files := conf.GetStringSlice("files")

	for _, f := range files {
		fh, err := b.Open(m, f)
		if err != nil {
			log.Printf("failed to open fortune file %q: %v", f, err)
			continue
		}
		m.fortunes[f] = getDict(fh)
		fh.Close()

		log.Printf("loaded fortune file %q", f)
	}

	return m, nil
}

func getDict(r io.Reader) []string {
	var dict []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		dict = append(dict, sc.Text())
	}
	return dict
}

type fortune struct {
	fortunes map[string][]string
}

func (f *fortune) Close() error { return nil }
func (f *fortune) Name() string { return mod.Name }

func (f *fortune) GetHandlers() map[core.EventType]core.EventHandler {
	m := map[core.EventType]core.EventHandler{}

	for name := range f.fortunes {
		db := name
		et := core.EventCommand{db}
		m[et] = core.EventHandlerFunc(func(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
			rw.Message(e.Target, f.getFortune(db))
			return nil
		})
	}

	return m
}

func (f *fortune) getFortune(dict string) string {
	d := f.fortunes[dict]
	return d[rand.Intn(len(d))]
}
