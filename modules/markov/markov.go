package markov

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/mischief/glenda/core"
	andrey "github.com/mischief/glenda/markov"
)

var mod = &core.ModInfo{"markov", create}

func init() {
	core.RegisterModule(mod)
}

func create(b core.Bot, conf *viper.Viper) (core.Module, error) {
	m := &markov{}

	dbfile := filepath.Join(b.Path(m), "markov.db")

	c, err := andrey.NewChain(dbfile)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}

	m.chain = c

	return m, nil
}

type markov struct {
	chain *andrey.Chain
}

func (m *markov) Close() error {
	return nil
}

func (m *markov) Name() string {
	return mod.Name
}

func (m *markov) GetHandlers() map[core.EventType]core.EventHandler {
	evs := map[core.EventType]core.EventHandler{}

	ec := core.EventCommand{mod.Name}
	evs[ec] = core.EventHandlerFunc(m.markov)

	em := core.EventMessage{core.Message}
	evs[em] = core.EventHandlerFunc(m.ingest)

	return evs
}

func (m *markov) markov(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	log.Printf("generate %+v", e)

	rw.Message(e.Target, m.chain.Generate(10))
	return nil
}

// FilterFunc returns true if s should be discarded.
type filterFunc func(s string) bool

func nomagic(s string) bool {
	return unicode.IsSymbol(rune(s[0]))
}

func tooshort(s string) bool {
	f := strings.Fields(s)
	return len(f) < 4
}

var filters = []filterFunc{
	nomagic,
	tooshort,
}

func (m *markov) ingest(ctx context.Context, rw core.ResponseWriter, e *core.Event) error {
	log.Printf("ingest %+v", e)

	for _, r := range filters {
		if r(e.Args) {
			return nil
		}
	}

	m.chain.Build(strings.NewReader(e.Args))

	return nil
}
