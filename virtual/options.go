package virtual

import (
	"context"

	"github.com/insei/gerpo/types"
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

func WithBoolEqFilter(fn func(b *BoolEQFilterBuilder)) Option {
	boolEQBuilder := &BoolEQFilterBuilder{}
	fn(boolEQBuilder)
	return columnOptionFn(func(c *column) {
		boolEQBuilder.validate(c.base.Field)
		c.base.Filters.AddFilterFn(types.OperationEQ, func(ctx context.Context, value any) (string, bool) {
			var b bool
			v, ok := value.(*bool)
			if ok && v != nil {
				b = *v
			}
			if (ok && v == nil) || value == nil {
				nilSQLStr := boolEQBuilder.nilSQL(ctx)
				return nilSQLStr, false
			}
			vv, ok := value.(bool)
			if ok {
				b = vv
			}
			if b {
				trueSQLStr := boolEQBuilder.trueSQL(ctx)
				return trueSQLStr, false
			}
			falseSQLStr := boolEQBuilder.falseSQL(ctx)
			return falseSQLStr, false
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
