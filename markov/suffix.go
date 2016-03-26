package markov

import (
	"bytes"
	"encoding/json"

	"github.com/jmcvetta/randutil"
)

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
