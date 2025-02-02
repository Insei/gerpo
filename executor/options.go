package executor

import (
	"github.com/insei/gerpo/executor/cache"
)

type options struct {
	cacheSource cache.Source
}

type Option interface {
	apply(c *options)
}

// optionFunc is a type that implements the Option interface.
type optionFn func(c *options)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given *options.
func (f optionFn) apply(c *options) {
	f(c)
}

func WithCacheSource(source cache.Source) Option {
	return optionFn(func(o *options) {
		if source != nil {
			o.cacheSource = source
		}
	})
}
