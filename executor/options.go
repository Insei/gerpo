package executor

import (
	"github.com/insei/gerpo/cache"
	"github.com/insei/gerpo/logger"
)

type options struct {
	cacheBundle cache.ModelBundle
	log         logger.Logger
}

type Option interface {
	apply(c *options)
}

// optionFunc is a type that implements the Option interface.
type optionFn func(c *options)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given Column.
func (f optionFn) apply(c *options) {
	f(c)
}

func WithCacheBundle(bundle cache.ModelBundle) Option {
	return optionFn(func(o *options) {
		if bundle != nil {
			o.cacheBundle = bundle
		}
	})
}

func WithLogger(logger logger.Logger) Option {
	return optionFn(func(o *options) {
		if logger != nil {
			o.log = logger
		}
	})
}
