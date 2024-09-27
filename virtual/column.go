package virtual

import (
	"context"
	"slices"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/filter"
	"github.com/insei/gerpo/types"
)

type column struct {
	base *types.ColumnBase
}

func (c *column) GetFilterFn(operation filter.Operation) (func(value any) (string, bool, error), bool) {
	return c.base.Filters.GetFilterFn(operation)
}

func (c *column) IsAllowedAction(act types.AllowedAction) bool {
	return slices.Contains(c.base.AllowedActions, act)
}

func (c *column) ToSQL(ctx context.Context) string {
	return c.base.ToSQL(ctx)
}

func (c *column) GetPtr(model any) any {
	return c.base.GetPtr(model)
}

func New(field fmap.Field, opts ...Option) types.Column {
	base := types.NewColumnBase(field, nil)
	c := &column{
		base: base,
	}
	c.base.AllowedActions = []types.AllowedAction{types.ActionRead}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}
