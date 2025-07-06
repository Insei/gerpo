package ctx

import (
	"github.com/insei/gerpo/logger"
)

type ctxStorageOption func(c *CtxCache)

// apply implements the Option interface for ctxStorageOption.
// It calls the underlying function with the given *CtxCache.
func (f ctxStorageOption) apply(c *CtxCache) {
	f(c)
}

type Option interface {
	apply(c *CtxCache)
}

func WithLogger(log logger.Logger) Option {
	return ctxStorageOption(func(s *CtxCache) {
		if log != nil {
			s.log = log
		}
	})
}

// WithKey sets unique key for cache storage to store cache for you model in shared ctx store
// This option is not recommended in base use cases.
func WithKey(key string) Option {
	return ctxStorageOption(func(s *CtxCache) {
		if key != "" {
			s.key = key
		}
	})
}
