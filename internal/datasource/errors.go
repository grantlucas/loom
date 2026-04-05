package datasource

import "errors"

// ErrBdNotFound indicates the bd binary was not found in PATH.
var ErrBdNotFound = errors.New("bd executable not found")
