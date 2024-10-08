package query

import (
	"context"
	"fmt"
	"slices"

	"github.com/insei/gerpo/types"
)

type StringSQLBuilder struct {
	ctx           context.Context
	joinsSQL      string
	orderBuilder  *StringSQLOrderBuilder
	whereBuilder  *StringSQLWhereBuilder
	selectBuilder *StringSQLSelectBuilder
}

func (b *StringSQLBuilder) WhereBuilder() *StringSQLWhereBuilder {
	return b.whereBuilder
}

func (b *StringSQLBuilder) OrderBuilder() *StringSQLOrderBuilder {
	return b.orderBuilder
}

func (b *StringSQLBuilder) SelectBuilder() *StringSQLSelectBuilder {
	return b.selectBuilder
}

func NewStringSQLBuilder(ctx context.Context) *StringSQLBuilder {
	return &StringSQLBuilder{
		ctx: ctx,
		whereBuilder: &StringSQLWhereBuilder{
			ctx: ctx,
		},
		orderBuilder: &StringSQLOrderBuilder{
			ctx: ctx,
		},
		selectBuilder: &StringSQLSelectBuilder{
			ctx: ctx,
		},
	}
}

type StringSQLSelectBuilder struct {
	ctx     context.Context
	exclude []types.Column
	sql     string
}

func (b *StringSQLSelectBuilder) Select(cols ...types.Column) {
	for _, col := range cols {
		if slices.Contains(b.exclude, col) {
			continue
		}
		if b.sql != "" {
			b.sql += ", "
		}
		b.sql += col.ToSQL(b.ctx)
	}
}

func (b *StringSQLSelectBuilder) Exclude(cols ...types.Column) {
	for _, col := range cols {
		b.exclude = append(b.exclude, col)
	}
}

type StringSQLOrderBuilder struct {
	ctx context.Context
	sql string
}

func (b *StringSQLOrderBuilder) OrderBy(columnDirection string) *StringSQLOrderBuilder {
	if b.sql != "" {
		b.sql += ", "
	}
	b.sql += columnDirection
	return b
}

func (b *StringSQLOrderBuilder) OrderByColumn(col types.Column, direction types.OrderDirection) error {
	if col.IsAllowedAction(types.SQLActionSort) {
		if b.sql != "" {
			b.sql += ", "
		}
		b.sql += col.ToSQL(b.ctx) + " " + string(direction)
	}
	return nil
}

type StringSQLWhereBuilder struct {
	ctx    context.Context
	sql    string
	values []any
}

func (b *StringSQLWhereBuilder) ToSQL() (string, []any) {
	return b.sql, b.values
}

func (b *StringSQLWhereBuilder) StartGroup() {
	b.sql += "("
}
func (b *StringSQLWhereBuilder) EndGroup() {
	b.sql += ")"
}

func (b *StringSQLWhereBuilder) AND() *StringSQLWhereBuilder {
	b.sql += " AND "
	return b
}

func (b *StringSQLWhereBuilder) OR() *StringSQLWhereBuilder {
	b.sql += " OR "
	return b
}

func (b *StringSQLWhereBuilder) AppendSQL(sql string, values ...any) {
	b.sql += sql
	b.values = append(b.values, values...)
}

func (b *StringSQLWhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		return fmt.Errorf("for field %s whereSQL %s option is not available", cl.GetField().GetStructPath(), operation)
	}
	sql, appendValue, err := filterFn(b.ctx, val)
	if err != nil {
		return err
	}
	b.sql += sql
	if appendValue {
		b.values = append(b.values, val)
	}
	return nil
}
