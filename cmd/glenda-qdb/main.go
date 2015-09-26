package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/mischief/glenda/util"

	"github.com/koding/kite"
)

var (
	qdb *kite.Kite
	DBs map[string]QDB

	urlDBs = map[string]string{
		"theo": "http://code.9front.org/hg/plan9front/raw-file/tip/lib/theo",
		"rob":  "http://code.9front.org/hg/plan9front/raw-file/tip/lib/rob",
		"ken":  "http://code.9front.org/hg/plan9front/raw-file/tip/lib/ken",
		"rsc":  "http://code.9front.org/hg/plan9front/raw-file/tip/lib/rsc",

		"troll": "http://code.9front.org/hg/plan9front/raw-file/tip/lib/troll",
		"terry": "http://code.9front.org/hg/plan9front/raw-file/tip/lib/terry",
		"roa":   "http://code.9front.org/hg/plan9front/raw-file/tip/lib/roa",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	qdb = util.NewKite("qdb", "1.0.0")

	DBs = make(map[string]QDB)

	for name, url := range urlDBs {
		d, err := newURLDB(url)
		if err != nil {
			qdb.Log.Warning("Failed to load quote database %q: %v", name, err)
		}

		DBs[name] = d
	}

	qdb.HandleFunc("qdb.getrandom", handleGetRandom).DisableAuthentication()
	<-qdb.ServerCloseNotify()
}

func handleGetRandom(r *kite.Request) (result interface{}, err error) {
	// [["dbname]]"
	var args []string
	err = r.Args.One().Unmarshal(&args)
	if err != nil || len(args) != 1 {
		return nil, fmt.Errorf("invalid arguments")
	}

	dbname := args[0]

	db, ok := DBs[dbname]
	if !ok {
		return nil, fmt.Errorf("database %q not found", dbname)
	}

	q, err := db.GetRandom()
	if err != nil {
		return nil, err
	}

	return q, nil
}

type QDB interface {
	GetRandom() (quote string, err error)
}

// urlDB loads a fortune file over http when it starts, and serves quote from the cached data.
type URLDB struct {
	quotes []string
}

func newURLDB(url string) (QDB, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	quotes := strings.Split(string(b), "\n")

	return QDB(&URLDB{quotes}), nil
}

func (u *URLDB) GetRandom() (quote string, err error) {
	return u.quotes[rand.Intn(len(u.quotes))], nil
}
