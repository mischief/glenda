package markov

import (
	"bytes"
	"testing"
)

func TestPrefixKey(t *testing.T) {
	p := NewPrefix(3)

	words := []string{"foo", "bar", "baz"}

	for _, w := range words {
		p.Shift(w)
	}

	want := []byte("foo bar baz")
	str := p.Key()

	if !bytes.Equal(want, str) {
		t.Errorf("expected key %q got %q", want, str)
	}
}
