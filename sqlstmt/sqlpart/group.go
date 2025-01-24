package sqlpart

import (
	"context"
	"strings"

	"github.com/insei/gerpo/types"
)

type Group interface {
	GroupBy(cols ...types.Column)
}

type GroupBuilder struct {
	ctx context.Context
	sql string
}

func NewGroupBuilder(ctx context.Context) *GroupBuilder {
	return &GroupBuilder{ctx: ctx}
}

func (b *GroupBuilder) SQL() string {
	if strings.TrimSpace(b.sql) == "" {
		return ""
	}
	return " GROUP BY " + b.sql
}

func (b *GroupBuilder) GroupBy(cols ...types.Column) {
	for _, col := range cols {
		if !col.IsAllowedAction(types.SQLActionGroup) {
			continue
			//TODO: log
		}
		b.sql += col.ToSQL(b.ctx)
	}
}
