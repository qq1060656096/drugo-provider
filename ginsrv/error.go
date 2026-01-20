package ginsrv

import "errors"

var (
	ErrAppNotFound     = errors.New("ginsrv: app not found in context")
	ErrAppTypeMismatch = errors.New("ginsrv: app type mismatch")
)
