package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type GetFirstHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any)
	OrderBy() types.OrderTarget
}

type GetFirstApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
	Order() sqlpart.Order
}

type GetFirst[TModel any] struct {
	baseModel *TModel

	whereBuilder   *linq.WhereBuilder
	orderBuilder   *linq.OrderBuilder
	excludeBuilder *linq.ExcludeBuilder
}

func (h *GetFirst[TModel]) Exclude(fieldPointers ...any) {
	h.excludeBuilder.Exclude(fieldPointers...)
}

func (h *GetFirst[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *GetFirst[TModel]) OrderBy() types.OrderTarget {
	return h.orderBuilder
}

func (h *GetFirst[TModel]) HandleFn(qFns ...func(m *TModel, h GetFirstHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func (h *GetFirst[TModel]) Apply(applier GetFirstApplier) {
	h.excludeBuilder.Apply(applier)
	h.whereBuilder.Apply(applier)
	h.orderBuilder.Apply(applier)
}

func NewGetFirst[TModel any](baseModel *TModel) *GetFirst[TModel] {
	return &GetFirst[TModel]{
		baseModel: baseModel,

		whereBuilder:   linq.NewWhereBuilder(baseModel),
		excludeBuilder: linq.NewExcludeBuilder(baseModel),
		orderBuilder:   linq.NewOrderBuilder(baseModel),
	}
}
