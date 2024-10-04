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

func (c *column) GetFilterFn(operation types.Operation) (func(ctx context.Context, value any) (string, bool, error), bool) {
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

func New(field fmap.Field, opts ...Option) types.Column {
	base := types.NewColumnBase(field, nil, filter.NewForField(field))
	c := &column{
		base: base,
	}
	c.base.AllowedActions = []types.SQLAction{types.SQLActionSelect}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}
