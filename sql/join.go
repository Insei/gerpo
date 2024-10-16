package sql

import (
	"context"
)

type StringJoinBuilder struct {
	ctx   context.Context
	joins []func(ctx context.Context) string
}

func (b *StringJoinBuilder) JOIN(joinFn func(ctx context.Context) string) {
	b.joins = append(b.joins, joinFn)
}

func (b *StringJoinBuilder) SQL() string {
	sql := ""
	for _, j := range b.joins {
		sql += " " + j(b.ctx)
	}
	return sql
}
