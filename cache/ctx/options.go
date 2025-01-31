package ctx

import (
	"github.com/insei/gerpo/logger"
)

type ctxSourceOption func(c *CtxCache)

// apply implements the Option interface for ctxSourceOption.
// It calls the underlying function with the given *CtxCache.
func (f ctxSourceOption) apply(c *CtxCache) {
	f(c)
}

type Option interface {
	apply(c *CtxCache)
}

func WithLogger(log logger.Logger) Option {
	return ctxSourceOption(func(s *CtxCache) {
		if log != nil {
			s.log = log
		}
	})
}

func WithKey(key string) Option {
	return ctxSourceOption(func(s *CtxCache) {
		if key != "" {
			s.key = key
		}
	})
}
