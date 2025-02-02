package types

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrWrongConfiguration = errors.New("wrong configuration")
)
