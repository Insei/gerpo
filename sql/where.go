package sql

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/types"
)

type StringWhereBuilder struct {
	ctx    context.Context
	sql    string
	values []any
}

func (b *StringWhereBuilder) ToSQL() (string, []any) {
	return b.sql, b.values
}

func (b *StringWhereBuilder) StartGroup() {
	b.sql += "("
}
func (b *StringWhereBuilder) EndGroup() {
	b.sql += ")"
}

func (b *StringWhereBuilder) AND() {
	b.sql += " AND "
}

func (b *StringWhereBuilder) OR() {
	b.sql += " OR "
}

func (b *StringWhereBuilder) AppendSQL(sql string, values ...any) {
	b.sql += sql
	b.values = append(b.values, values...)
}

func (b *StringWhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
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
