package core

import (
	"io"
)

type File interface {
	io.Reader
	io.Writer
	io.Closer
}

type Bot interface {
	// Open a file for a module in the data dir.
	Open(m Module, name string) (File, error)
	// Path prefix for a module's data dir.
	Path(m Module) string
	// Get magic character
	Magic() string
	// Get our preferred name
	Nick() string
}
