package ctx

import (
	"github.com/insei/gerpo/logger"
)

type ctxSourceOption func(c *ctxSource)

// apply implements the Option interface for ctxSourceOption.
// It calls the underlying function with the given Column.
func (f ctxSourceOption) apply(c *ctxSource) {
	f(c)
}

type Option interface {
	apply(c *ctxSource)
}

func WithLogger(log logger.Logger) Option {
	return ctxSourceOption(func(s *ctxSource) {
		if log != nil {
			s.log = log
		}
	})
}

func WithKey(key string) Option {
	return ctxSourceOption(func(s *ctxSource) {
		if key != "" {
			s.key = key
		}
	})
}
