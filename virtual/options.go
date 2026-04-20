package virtual

import (
	"context"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// columnOptionFn is a type that implements the VirtualOption interface.
type columnOptionFn func(c *column)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given *column.
func (f columnOptionFn) apply(c *column) {
	f(c)
}

type Option interface {
	apply(c *column)
}

// WithCompute sets a static SQL expression for the virtual column. The expression
// is always wrapped in parentheses — that is part of the contract, not magic — so
// it composes cleanly inside larger predicates. Optional bound args travel with the
// column wherever it is referenced (SELECT/WHERE/ORDER).
//
// When the column is not Aggregate, the standard set of operators (EQ, LT, IN, ...)
// is auto-derived from the field type, the same way it works for plain columns.
func WithCompute(sql string, args ...any) Option {
	wrapped := "(" + sql + ")"
	return columnOptionFn(func(c *column) {
		c.base.ToSQL = func(context.Context) string { return wrapped }
		if len(args) > 0 {
			c.base.SQLArgs = append([]any(nil), args...)
		}
		if !c.base.IsAggregate() {
			for op, fn := range sqlpart.GetFieldTypeFilters(c.base.Field, wrapped) {
				c.base.Filters.AddFilterFn(op, fn)
			}
		}
	})
}

// WithAggregate marks the column as an aggregate expression. Aggregate columns
// cannot be filtered in WHERE without an explicit Filter override — auto-derived
// operators are intentionally not registered.
func WithAggregate() Option {
	return columnOptionFn(func(c *column) {
		c.base.Aggregate = true
	})
}

// WithFilter registers a custom filter for one operation. spec is a FilterSpec —
// see virtual.SQL / Bound / SQLArgs / Match / Func. Other operators keep their
// auto-derived implementations (unless the column is Aggregate).
func WithFilter(op types.Operation, spec FilterSpec) Option {
	fn := compileFilter(spec)
	return columnOptionFn(func(c *column) {
		c.base.Filters.AddFilterFnArgs(op, fn)
		if c.base.FilterOverrides == nil {
			c.base.FilterOverrides = map[types.Operation]bool{}
		}
		c.base.FilterOverrides[op] = true
	})
}
