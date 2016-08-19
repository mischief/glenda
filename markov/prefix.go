package markov

import (
	"strings"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

func NewPrefix(length int) Prefix {
	sl := make([]string, length)
	return Prefix(sl)
}

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
	return []byte(p.String())
}
