package markov

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"

	"github.com/boltdb/bolt"
)

// order 3 markov chain stored in boltdb.

const (
	markovOrder = 3
)

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	db *bolt.DB

	// markov bucket name
	bucket []byte
	// prefix bucket name
	prefix []byte

	// in-memory prefix store
	prefixes map[string]bool
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(dbpath string) (*Chain, error) {
	c := &Chain{}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}

	c.db = db

	c.bucket = []byte("markov")
	c.prefix = []byte("prefix")

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(c.bucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(c.prefix)
		if err != nil {
			return err
		}

		return nil
	})

	c.prefixes, _ = c.getprefixes()

	if err != nil {
		db.Close()
		return nil, err
	}

	return c, nil
}

func (c *Chain) Close() error {
	c.prefixes = map[string]bool{}
	return c.db.Close()
}

func (c *Chain) upgram(pre Prefix, suf *Suffix) error {
	k := pre.Key()
	v := suf.Value()

	err := c.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(c.bucket).Put(k, v)
		return err
	})

	return err
}

func (c *Chain) getgram(pre Prefix) *Suffix {
	k := pre.Key()
	s := NewSuffix()

	c.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(c.bucket).Get(k)
		if v != nil {
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}
		}

		return nil
	})

	return s
}

// write a sentence prefix into the db
func (c *Chain) putprefix(pre Prefix) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		bu := tx.Bucket(c.prefix)
		seq, err := bu.NextSequence()
		if err != nil {
			return err
		}

		k := []byte(fmt.Sprintf("%d", seq))
		v := pre.Key()
		return tx.Bucket(c.prefix).Put(k, v)
	})

	return err
}

func (c *Chain) getprefixes() (map[string]bool, error) {
	r := make(map[string]bool)

	err := c.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket(c.prefix)

		return bu.ForEach(func(k, v []byte) error {
			r[string(v)] = true
			return nil
		})
	})

	return r, err
}

func scanSentence(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexAny(data, ".?!"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) error {
	scans := bufio.NewScanner(r)
	scans.Split(scanSentence)

	for scans.Scan() {
		sent := strings.ToLower(scans.Text())
		sentr := strings.NewReader(sent)

		scw := bufio.NewScanner(sentr)
		scw.Split(bufio.ScanWords)

		words := 0
		first := true

		p := make(Prefix, markovOrder)

		for scw.Scan() {
			word := scw.Text()
			words++
			if words < markovOrder+1 {
				p.Shift(word)
				continue
			}

			if first {
				c.prefixes[p.String()] = true
				c.putprefix(p)
				first = false
			}

			suf := c.getgram(p)
			suf.Insert(word)
			c.upgram(p, suf)

			p.Shift(word)
		}
	}

	return nil
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	nprefix := len(c.prefixes)
	if nprefix == 0 {
		return ""
	}

	if n < markovOrder {
		return ""
	}

	var p Prefix

	rnd := rand.Intn(nprefix)
	i := 0
	for k, _ := range c.prefixes {
		if i == rnd {
			p = Prefix(strings.Split(k, " "))
			break
		}
		i++
	}

	var words []string

	words = append(words, p...)

	for {
		suf := c.getgram(p)
		if len(suf.M) == 0 {
			break
		}

		word := suf.Pick()

		words = append(words, word)

		if strings.Contains(word, ".?!") {
			break
		}
		p.Shift(word)
	}

	return strings.Join(words, " ")
}

func (c *Chain) Dump() {
	log.Printf("%d prefixes", len(c.prefixes))
	for k, _ := range c.prefixes {
		log.Printf("- %+v", k)
	}
}
