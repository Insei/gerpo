package sql

import (
	"context"

	"github.com/insei/gerpo/types"
)

type StringOrderBuilder struct {
	ctx context.Context
	sql string
}

func (b *StringOrderBuilder) OrderBy(columnDirection string) *StringOrderBuilder {
	if b.sql != "" {
		b.sql += ", "
	}
	b.sql += columnDirection
	return b
}

func (b *StringOrderBuilder) OrderByColumn(col types.Column, direction types.OrderDirection) error {
	if col.IsAllowedAction(types.SQLActionSort) {
		if b.sql != "" {
			b.sql += ", "
		}
		b.sql += col.ToSQL(b.ctx) + " " + string(direction)
	}
	return nil
}
