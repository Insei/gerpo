package ctx

import (
	"github.com/insei/gerpo/logger"
)

type cacheOption func(c *Cache)

// apply implements the Option interface for cacheOption.
// It calls the underlying function with the given *Cache.
func (f cacheOption) apply(c *Cache) {
	f(c)
}

type Option interface {
	apply(c *Cache)
}

func WithLogger(log logger.Logger) Option {
	return cacheOption(func(s *Cache) {
		if log != nil {
			s.log = log
		}
	})
}

// WithKey sets unique key for cache storage to store cache for you model in shared ctx store
// This option is not recommended in base use cases.
func WithKey(key string) Option {
	return cacheOption(func(s *Cache) {
		if key != "" {
			s.key = key
		}
	})
}
