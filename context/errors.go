package context

import (
	"errors"
)

var (
	ErrFunctionClosed = errors.New("function is closed")
	ErrStreamClosed   = errors.New("stream is closed")
)

var (
	ErrUndefinedValue    = errors.New("undefined value")
	ErrUndefinedFunction = errors.New("undefined function")
	ErrClosedChannel     = errors.New("closed channel")
)
