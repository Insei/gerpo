package sql

import (
	"context"
	"slices"
	"strconv"

	"github.com/insei/gerpo/types"
)

type StringSelectBuilder struct {
	ctx     context.Context
	exclude []types.Column
	columns []types.Column
	limit   uint64
	offset  uint64
}

func (b *StringSelectBuilder) Select(cols ...types.Column) {
	for _, col := range cols {
		b.columns = append(b.columns, col)
	}
}

func (b *StringSelectBuilder) Exclude(cols ...types.Column) {
	for _, col := range cols {
		b.exclude = append(b.exclude, col)
	}
}

func (b *StringSelectBuilder) Limit(limit uint64) {
	b.limit = limit
}

func (b *StringSelectBuilder) Offset(offset uint64) {
	b.offset = offset
}

func (b *StringSelectBuilder) GetColumns() []types.Column {
	cols := make([]types.Column, 0, len(b.columns))
	for _, col := range b.columns {
		if slices.Contains(b.exclude, col) {
			//TODO: log
			continue
		}
		cols = append(cols, col)
	}
	return cols
}

func (b *StringSelectBuilder) GetSQL() string {
	sql := ""
	for _, col := range b.columns {
		if slices.Contains(b.exclude, col) {
			continue
		}
		if sql != "" {
			sql += ", "
		}
		sql += col.ToSQL(b.ctx)
	}
	return sql
}
func (b *StringSelectBuilder) GetLimit() string {
	if b.limit == 0 {
		return ""
	}
	return strconv.FormatUint(b.limit, 10)
}

func (b *StringSelectBuilder) GetOffset() string {
	if b.offset == 0 {
		return ""
	}
	return strconv.FormatUint(b.offset, 10)
}
