package markov

import (
	"bytes"
	"testing"
)

func TestSuffixInsert(t *testing.T) {
	s := NewSuffix()

	word := "foo"
	count := 100

	for i := 0; i < count; i++ {
		s.Insert(word)
	}

	if s.M[word] != uint32(count) {
		t.Errorf("expected count %d got %d", count, s.M[word])
	}
}

func TestSuffixPick(t *testing.T) {
	s := NewSuffix()

	word := "foo"
	count := 1

	for i := 0; i < count; i++ {
		s.Insert(word)
	}

	w := s.Pick()

	if w != word {
		t.Errorf("expected string %q got %q", word, w)
	}
}

func TestSuffixMerge(t *testing.T) {
	s := NewSuffix()

	word := "foo"
	count := 1

	for i := 0; i < count; i++ {
		s.Insert(word)
	}

	w := s.Pick()

	if w != word {
		t.Errorf("expected string %q got %q", word, w)
	}

	// use merge as a copy
	s2 := NewSuffix()

	s2.Merge(s)

	if s2.M[word] != s.M[word] {
		t.Errorf("merge is broken")
	}
}

func TestSuffixValue(t *testing.T) {
	s := NewSuffix()

	word := "foo"
	count := 1

	for i := 0; i < count; i++ {
		s.Insert(word)
	}

	// seems Value generates a newline. shrug.
	want := []byte(`{"M":{"foo":1}}
`)
	val := s.Value()

	if !bytes.Equal(val, want) {
		t.Errorf("expected value (%d) %s got (%d) %s", len(want), want, len(val), val)
	}

}
