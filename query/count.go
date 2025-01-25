package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type CountHelper[TModel any] interface {
	Where() types.WhereTarget
}

type CountApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Where() sqlpart.Where
}

type Count[TModel any] struct {
	baseModel any

	whereBuilder *linq.WhereBuilder
	groupBuilder *linq.GroupBuilder
	joinBuilder  *linq.JoinBuilder
}

func (h *Count[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Count[TModel]) Apply(applier CountApplier) {
	h.whereBuilder.Apply(applier)
}

func (h *Count[TModel]) HandleFn(qFns ...func(m *TModel, h CountHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel.(*TModel), h)
	}
}

func NewCount[TModel any](baseModel *TModel) *Count[TModel] {
	return &Count[TModel]{
		baseModel:    baseModel,
		whereBuilder: linq.NewWhereBuilder(baseModel),
		groupBuilder: linq.NewGroupBuilder(baseModel),
		joinBuilder:  linq.NewJoinBuilder(),
	}
}
