// Package types defines the interfaces a cache backend must satisfy to plug into gerpo's
// executor as well as the canonical sentinel errors. The "types" name is kept for
// backwards compatibility with the public API.
package types //nolint:revive // public API package name kept for backwards compatibility

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrWrongConfiguration = errors.New("wrong configuration")
)
