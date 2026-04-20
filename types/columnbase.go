package types

import (
	"context"
	"slices"

	"github.com/insei/fmap/v3"
)

// NewColumnBase creates a new instance of ColumnBase with the provided field, SQL generation function, and filters manager.
func NewColumnBase(field fmap.Field, toSQLFn func(ctx context.Context) string, filters SQLFilterManager) *ColumnBase {
	return &ColumnBase{
		Field:          field,
		ToSQL:          toSQLFn,
		AllowedActions: make([]SQLAction, 0),
		Filters:        filters,
		GetPtr: func(model any) any {
			return field.GetPtr(model)
		},
	}
}

// ColumnBase represents a base structure for defining and manipulating column-related behaviors in SQL operations.
type ColumnBase struct {
	Field          fmap.Field
	ToSQL          func(ctx context.Context) string
	AllowedActions []SQLAction
	Filters        SQLFilterManager
	GetPtr         func(model any) any

	// SQLArgs holds bound parameters that belong to the column expression itself
	// (e.g. virtual columns declared via Compute(sql, args...)). They are merged
	// into the bound-args list whenever the column appears in SELECT/WHERE/ORDER.
	SQLArgs []any

	// Aggregate marks the column as an aggregate expression (SUM, COUNT, ...).
	// Aggregate columns cannot be used in WHERE without an explicit Filter() override
	// — WhereBuilder rejects them with a clear error.
	Aggregate bool

	// FilterOverrides records operations whose filter was registered explicitly
	// (e.g. virtual.Filter(op, spec)), so the WHERE-builder can distinguish
	// auto-derived filters from user-provided ones for aggregate gating.
	FilterOverrides map[Operation]bool
}

// IsAllowedAction determines if a given SQLAction is allowed for the column.
func (c *ColumnBase) IsAllowedAction(act SQLAction) bool {
	return slices.Contains(c.AllowedActions, act)
}

// IsAggregate reports whether the column was marked as an aggregate expression.
func (c *ColumnBase) IsAggregate() bool {
	return c.Aggregate
}

// HasFilterOverride reports whether a custom filter was registered for the operation.
func (c *ColumnBase) HasFilterOverride(op Operation) bool {
	if c.FilterOverrides == nil {
		return false
	}
	return c.FilterOverrides[op]
}
