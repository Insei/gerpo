package sql

import (
	"context"

	"github.com/insei/gerpo/types"
)

type StringGroupBuilder struct {
	ctx context.Context
	sql string
}

func (b *StringGroupBuilder) GroupBy(cols ...types.Column) error {
	for _, col := range cols {
		if !col.IsAllowedAction(types.SQLActionGroup) {
			continue
			//TODO: log
		}
		b.sql += col.ToSQL(b.ctx)
	}
	return nil
}
