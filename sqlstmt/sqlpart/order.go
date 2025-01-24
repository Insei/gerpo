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
	orderBy string
	ctx     context.Context
}

func NewOrderBuilder(ctx context.Context) *OrderBuilder {
	return &OrderBuilder{
		ctx: ctx,
	}
}

func (b *OrderBuilder) OrderBy(columnAndDirection string) {
	if b.orderBy != "" {
		b.orderBy += ", "
	}
	b.orderBy += columnAndDirection
}

func (b *OrderBuilder) OrderByColumn(col types.Column, direction types.OrderDirection) {
	if !col.IsAllowedAction(types.SQLActionSort) {
		//TODO: Log it
		return
	}
	if b.orderBy != "" {
		b.orderBy += ", "
	}
	b.orderBy += col.ToSQL(b.ctx) + " " + string(direction)
}

func (b *OrderBuilder) SQL() string {
	if strings.TrimSpace(b.orderBy) == "" {
		return ""
	}
	return " ORDER BY " + b.orderBy
}
