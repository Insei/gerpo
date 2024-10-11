package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type CountUserHelper[TModel any] interface {
	Where() types.WhereTarget
}

type CountHelper[TModel any] interface {
	CountUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h CountUserHelper[TModel]))
}

type countHelper[TModel any] struct {
	core              *linq.CoreBuilder
	excludeBuilder    *linq.ExcludeBuilder
	paginationBuilder *linq.PaginationBuilder
	whereBuilder      *linq.WhereBuilder
	orderBuilder      *linq.OrderBuilder
}

func (h *countHelper[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *countHelper[TModel]) Apply(sqlBuilder *sql.StringBuilder) {
	h.excludeBuilder.Apply(sqlBuilder.SelectBuilder())
	h.paginationBuilder.Apply(sqlBuilder.SelectBuilder())
	h.orderBuilder.Apply(sqlBuilder.SelectBuilder())
	h.whereBuilder.Apply(sqlBuilder.WhereBuilder())
}

func (h *countHelper[TModel]) HandleFn(qFns ...func(m *TModel, h CountUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func newCountHelper[TModel any](core *linq.CoreBuilder) *countHelper[TModel] {
	paginationBuilder := linq.NewPaginationBuilder()
	return &countHelper[TModel]{
		core:              core,
		excludeBuilder:    linq.NewExcludeBuilder(core, types.SQLActionSelect),
		paginationBuilder: paginationBuilder,
		whereBuilder:      linq.NewWhereBuilder(core),
		orderBuilder:      linq.NewOrderBuilder(core),
	}
}

func NewCountHelper[TModel any](core *linq.CoreBuilder) CountHelper[TModel] {
	return newCountHelper[TModel](core)
}
