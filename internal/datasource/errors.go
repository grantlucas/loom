package datasource

import (
	"errors"
	"fmt"
)

// ErrBdNotFound indicates the bd binary was not found in PATH.
var ErrBdNotFound = errors.New("bd executable not found")

// ErrProjectNotInitialized indicates no .beads directory was found.
var ErrProjectNotInitialized = errors.New("no beads project found")

// ErrMalformedResponse indicates bd returned output that could not be parsed.
var ErrMalformedResponse = errors.New("malformed response from bd")

// BdError wraps a non-zero exit from the bd binary with stderr context.
type BdError struct {
	Args   []string
	Stderr string
	Err    error
}

func (e *BdError) Error() string {
	return fmt.Sprintf("bd %v: %s", e.Args, e.Stderr)
}

func (e *BdError) Unwrap() error {
	return e.Err
}
