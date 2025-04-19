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
	sql strings.Builder
}

func NewGroupBuilder(ctx context.Context) *GroupBuilder {
	return &GroupBuilder{ctx: ctx}
}

func (b *GroupBuilder) SQL() string {
	if b.sql.Len() < 1 {
		return ""
	}
	return " GROUP BY " + b.sql.String()
}

func (b *GroupBuilder) GroupBy(cols ...types.Column) {
	for _, col := range cols {
		if !col.IsAllowedAction(types.SQLActionGroup) {
			continue
			//TODO: error
		}
		sql := strings.TrimSpace(col.ToSQL(b.ctx))
		if len(sql) < 1 {
			continue
			//TODO: error
		}
		b.sql.WriteString(sql)
	}
}
