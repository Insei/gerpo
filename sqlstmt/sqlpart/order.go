package sqlpart

import (
	"context"
	"strings"

	"github.com/insei/gerpo/types"
)

type Order interface {
	OrderBy(columnAndDirection string)
	OrderByColumn(col types.Column, direction types.OrderDirection)
}

type OrderBuilder struct {
	orderBy strings.Builder
	ctx     context.Context
}

func NewOrderBuilder(ctx context.Context) *OrderBuilder {
	return &OrderBuilder{
		ctx: ctx,
	}
}

// Reset prepares the builder for reuse by a new query without dropping the underlying buffer.
func (b *OrderBuilder) Reset(ctx context.Context) {
	b.ctx = ctx
	b.orderBy.Reset()
}

func (b *OrderBuilder) OrderBy(columnAndDirection string) {
	if b.orderBy.Len() > 0 {
		b.orderBy.WriteString(", ")
	}
	b.orderBy.WriteString(columnAndDirection)
}

func (b *OrderBuilder) OrderByColumn(col types.Column, direction types.OrderDirection) {
	if !col.IsAllowedAction(types.SQLActionSort) {
		//TODO: error
		return
	}
	sql := col.ToSQL(b.ctx)
	if len(sql) < 1 {
		//TODO: error
		return
	}
	if b.orderBy.Len() > 0 {
		b.orderBy.WriteString(", ")
	}
	b.orderBy.WriteString(sql)
	b.orderBy.WriteString(" ")
	b.orderBy.WriteString(string(direction))
}

func (b *OrderBuilder) SQL() string {
	if b.orderBy.Len() < 1 {
		return ""
	}
	return " ORDER BY " + b.orderBy.String()
}
