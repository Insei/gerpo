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
}

// IsAllowedAction determines if a given SQLAction is allowed for the column.
func (c *ColumnBase) IsAllowedAction(act SQLAction) bool {
	return slices.Contains(c.AllowedActions, act)
}
