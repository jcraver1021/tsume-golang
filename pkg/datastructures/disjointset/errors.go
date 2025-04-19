package disjointset

import (
	"errors"
)

var (
	ErrInvalidConfiguration = errors.New("configuration is invalid")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrMissingElement       = errors.New("element not found in disjoint set")
)
