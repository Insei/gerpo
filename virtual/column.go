package virtual

import (
	"context"
	"fmt"
	"slices"

	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/types"
)

type column struct {
	base *types.ColumnBase
}

func (c *column) GetAvailableFilterOperations() []types.Operation {
	return c.base.Filters.GetAvailableFilterOperations()
}

func (c *column) IsAvailableFilterOperation(operation types.Operation) bool {
	return c.base.Filters.IsAvailableFilterOperation(operation)
}

func (c *column) GetFilterFn(operation types.Operation) (func(ctx context.Context, value any) (string, []any, error), bool) {
	return c.base.Filters.GetFilterFn(operation)
}

func (c *column) IsAllowedAction(act types.SQLAction) bool {
	return slices.Contains(c.base.AllowedActions, act)
}

func (c *column) ToSQL(ctx context.Context) string {
	return c.base.ToSQL(ctx)
}

func (c *column) GetPtr(model any) any {
	return c.base.GetPtr(model)
}

func (c *column) GetField() fmap.Field {
	return c.base.Field
}

func (c *column) GetAllowedActions() []types.SQLAction {
	return c.base.AllowedActions
}
func (c *column) Name() (string, bool) {
	return "", false
}

func (c *column) Table() (string, bool) {
	return "", false
}

func (c *column) IsAggregate() bool {
	return c.base.IsAggregate()
}

func (c *column) HasFilterOverride(op types.Operation) bool {
	return c.base.HasFilterOverride(op)
}

// SQLArgs returns the bound parameters declared via Compute(sql, args...). They
// must be appended to the bound list whenever the column expression is referenced
// in SELECT or in an auto-derived WHERE filter.
func (c *column) SQLArgs() []any {
	return c.base.SQLArgs
}

// IsReturned reports whether the column should appear in a RETURNING clause for
// the given action. Virtual columns are SELECT-only, so the answer is always false
// — including a virtual expression in RETURNING does not make sense.
func (c *column) IsReturned(_ types.SQLAction) bool {
	return false
}

func New(field fmap.Field, opts ...Option) (types.Column, error) {
	if field == nil {
		return nil, fmt.Errorf("field is nil")
	}
	base := types.NewColumnBase(field, nil, types.NewFilterManagerForField(field))
	c := &column{
		base: base,
	}
	c.base.AllowedActions = []types.SQLAction{types.SQLActionSelect}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c, nil
}
