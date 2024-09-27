package virtual

import (
	"context"

	"github.com/insei/gerpo/filter"
)

// columnOptionFn is a type that implements the VirtualOption interface.
type columnOptionFn func(c *column)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given Column.
func (f columnOptionFn) apply(c *column) {
	f(c)
}

type Option interface {
	apply(c *column)
}

func WithBoolEqFilter(trueSQL, falseSQL, nilSQL string) Option {
	return columnOptionFn(func(c *column) {
		c.base.Filters.AddFilterFn(filter.OperationEQ, func(value any) (string, bool) {
			var b bool
			v, ok := value.(*bool)
			if ok && v != nil {
				b = *v
			}
			if v == nil || value == nil {
				return nilSQL, false
			}
			vv, ok := value.(bool)
			if ok {
				b = vv
			}
			if b {
				return trueSQL, false
			}
			return falseSQL, false
		})
	})
}

func WithSQL(fn func(ctx context.Context) string) Option {
	return columnOptionFn(func(c *column) {
		if fn != nil {
			c.base.ToSQL = fn
		}
	})
}
