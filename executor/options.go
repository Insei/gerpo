package executor

import "github.com/insei/gerpo/cache"

type options struct {
	cacheBundle cache.ModelBundle
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
