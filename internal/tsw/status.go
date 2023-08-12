package tsw

import "errors"

type Status struct {
	Success bool
	Message string
}

var ErrNotFound = errors.New("not found")
