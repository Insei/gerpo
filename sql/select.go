package sql

import (
	"context"
	"slices"
	"strconv"

	"github.com/insei/gerpo/types"
)

type StringSelectBuilder struct {
	ctx     context.Context
	columns []types.Column
	limit   uint64
	offset  uint64
	orderBy string
}

// deleteFunc Modified deleteFunc from slices packages without clean element,
// removes any elements from s for which del returns true,
// returning the modified slice.
// deleteFunc zeroes the elements between the new length and the original length.
func deleteFunc[S ~[]E, E any](s S, del func(E) bool) S {
	i := slices.IndexFunc(s, del)
	if i == -1 {
		return s
	}
	var newSlice []E = make([]E, 0, len(s))
	// Don't start copying elements until we find one to delete.
	for j := i + 1; j < len(s); j++ {
		if v := s[j]; !del(v) {
			newSlice = append(newSlice, v)
		}
	}
	//clear(s[i:]) // zero/nil out the obsolete elements, for GC
	return newSlice
}

func (b *StringSelectBuilder) Exclude(cols ...types.Column) {
	b.columns = deleteFunc(b.columns, func(column types.Column) bool {
		if slices.Contains(cols, column) {
			return true
		}
		return false
	})
}

func (b *StringSelectBuilder) Limit(limit uint64) {
	b.limit = limit
}

func (b *StringSelectBuilder) Offset(offset uint64) {
	b.offset = offset
}

func (b *StringSelectBuilder) OrderBy(columnDirection string) *StringSelectBuilder {
	if b.orderBy != "" {
		b.orderBy += ", "
	}
	b.orderBy += columnDirection
	return b
}

func (b *StringSelectBuilder) OrderByColumn(col types.Column, direction types.OrderDirection) error {
	if col.IsAllowedAction(types.SQLActionSort) {
		if b.orderBy != "" {
			b.orderBy += ", "
		}
		b.orderBy += col.ToSQL(b.ctx) + " " + string(direction)
	}
	return nil
}

func (b *StringSelectBuilder) GetColumns() []types.Column {
	return b.columns
}

func (b *StringSelectBuilder) GetSQL() string {
	sql := ""
	for _, col := range b.columns {
		if sql != "" {
			sql += ", "
		}
		sql += col.ToSQL(b.ctx)
	}
	return sql
}

func (b *StringSelectBuilder) GetOrderSQL() string {
	return b.orderBy
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
