package column

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/filter"
	"github.com/insei/gerpo/types"
)

type column struct {
	query string
	base  *types.ColumnBase
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

func generateSQLQuery(opt *options) string {
	sql := opt.name
	if opt.table != "" {
		sql = fmt.Sprintf("%s.%s", strings.TrimSpace(opt.table), sql)
	}
	return sql
}

func generateToSQLFn(sql, alias string) func(ctx context.Context) string {
	if len(alias) > 0 {
		sql += " AS " + strings.TrimSpace(alias)
	}
	return func(ctx context.Context) string {
		return sql
	}
}

func New(field fmap.Field, opts ...Option) types.Column {
	forOpts := &options{}
	forOpts.name = strings.TrimSpace(toSnakeCase(field.GetName()))
	for _, opt := range opts {
		opt.apply(forOpts)
	}

	query := generateSQLQuery(forOpts)
	base := types.NewColumnBase(field, generateToSQLFn(query, forOpts.alias), filter.NewForField(field))
	c := &column{
		base:  base,
		query: query,
	}
	filters := filter.GetAvailableFilters(field, query)
	for op, filterFn := range filters {
		c.base.Filters.AddFilterFn(op, filterFn)
	}
	c.base.AllowedActions = []types.SQLAction{types.SQLActionInsert, types.SQLActionSelect, types.SQLActionUpdate, types.SQLActionSort}
	c.base.AllowedActions = slices.DeleteFunc(c.base.AllowedActions, func(action types.SQLAction) bool {
		return slices.Contains(forOpts.notAvailActions, action)
	})
	return c
}
