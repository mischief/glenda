package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/jmcvetta/randutil"
	"github.com/kballard/goirc/irc"
)

// order 3 markov chain stored in boltdb.

const (
	markovOrder = 3
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// convert prefix into a byte array for db
func (p Prefix) Key() []byte {
	buf := new(bytes.Buffer)
	io.WriteString(buf, strings.Join(p, " "))
	return buf.Bytes()
}

type Suffix struct {
	M map[string]uint32
}

func NewSuffix() *Suffix {
	s := &Suffix{}
	s.M = make(map[string]uint32)
	return s
}

func (s *Suffix) Insert(word string) {
	if c, ok := s.M[word]; ok {
		c++
		s.M[word] = c
		return
	}

	s.M[word] = 1
}

func (s *Suffix) Pick() string {
	var c []randutil.Choice

	for k, v := range s.M {
		c = append(c, randutil.Choice{int(v), k})
	}

	choice, err := randutil.WeightedChoice(c)
	if err != nil {
		panic(err)
	}

	return choice.Item.(string)
}

func (s *Suffix) Merge(other *Suffix) {
	for k, v := range other.M {
		s.M[k] += v
	}
}

// convert suffix into json encoded value for db
func (s *Suffix) Value() []byte {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(s); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

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

func init() {
	RegisterModule("markov", func() Module {
		return &MarkovMod{}
	})
}

type MarkovMod struct {
	chain *Chain
}

func (m *MarkovMod) Init(b *Bot, conn irc.SafeConn) error {
	//conf := b.Config.Search("mod", "markov")

	c, err := NewChain(filepath.Join(b.DataDir, "markov"))
	if err != nil {
		return fmt.Errorf("error opening db: %s", err)
	}

	m.chain = c

	b.Hook("markov", func(b *Bot, sender, cmd string, args ...string) error {
		b.Conn.Privmsg(sender, m.chain.Generate(rand.Intn(10)+10))
		return nil
	})

	conn.AddHandler("PRIVMSG", func(c *irc.Conn, l irc.Line) {
		if strings.HasPrefix(l.Args[0], b.Magic) {
			return
		}

		m.chain.Build(strings.NewReader(l.Args[1]))
	})

	log.Printf("markov module initialized")
	return nil
}

func (m *MarkovMod) Reload() error {
	return nil
}

func (m *MarkovMod) Call(args ...string) error {
	return nil
}
