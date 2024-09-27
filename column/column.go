package column

import (
	"context"
	"fmt"
	"reflect"
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

func generateQuerySQL(opt *options) string {
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

	query := generateQuerySQL(forOpts)
	base := types.NewColumnBase(field, generateToSQLFn(query, forOpts.alias))
	c := &column{
		base:  base,
		query: query,
	}
	c.base.AllowedActions = []types.AllowedAction{types.ActionRead, types.ActionUpdate, types.ActionSort}
	return c
}

func genEQFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		if value == nil {
			return fmt.Sprintf("%s IS NULL", query), true
		}
		return fmt.Sprintf("%s = (?)", query), true
	}
}

func genNEQFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		if value == nil {
			return fmt.Sprintf("%s IS NOT NULL", query), true
		}
		return fmt.Sprintf("%s != (?)", query), true
	}
}

func genLTFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		return fmt.Sprintf("%s < (?)", query), true
	}
}
func genLTEFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		return fmt.Sprintf("%s <= (?)", query), true
	}
}

func genGTFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		return fmt.Sprintf("%s > (?)", query), true
	}
}
func genGTEFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		return fmt.Sprintf("%s >= (?)", query), true
	}
}

func genINFn(query string) func(value any) (string, bool) {
	return func(value any) (string, bool) {
		reflect.Va
		return fmt.Sprintf("%s IN (?)", query), true
	}
}

func getAvailableFilters(field fmap.Field, query string) map[filter.Operation]func(value any) (string, bool) {
	filters := make(map[filter.Operation]func(value any) (string, bool))
	eqFn := genEQFn(query)
	neqFn := genNEQFn(query)
	ltFn := genLTFn(query)
	lteFn := genLTEFn(query)
	gtFn := genGTFn(query)
	gteFn := genGTEFn(query)
	derefType := field.GetDereferencedType()
	switch derefType.Kind() {
	case reflect.Bool:
		filters[filter.OperationEQ] = eqFn
		filters[filter.OperationNEQ] = neqFn
	case reflect.String:
		filters[filter.OperationEQ] = eqFn
		filters[filter.OperationNEQ] = neqFn
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		filters[filter.OperationEQ] = eqFn
		filters[filter.OperationNEQ] = neqFn
		filters[filter.OperationLT] = ltFn
		filters[filter.OperationLTE] = lteFn
		filters[filter.OperationGT] = gtFn
		filters[filter.OperationGTE] = gteFn
	default:
		panic("unhandled default case")

	}
}
