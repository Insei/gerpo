package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// DeleteHelper is an interface that provides functionality for constructing delete queries using conditions.
type DeleteHelper[TModel any] interface {
	// Where defines the starting point for building conditions in a query, returning a types.WhereTarget interface.
	Where() types.WhereTarget
}

type DeleteApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Where() sqlpart.Where
	Join() sqlpart.Join
}

type Delete[TModel any] struct {
	baseModel *TModel

	whereBuilder *linq.WhereBuilder
	joinBuilder  *linq.JoinBuilder
}

func (h *Delete[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Delete[TModel]) Apply(applier DeleteApplier) {
	h.whereBuilder.Apply(applier)
	h.joinBuilder.Apply(applier)
}

func (h *Delete[TModel]) HandleFn(qFns ...func(m *TModel, h DeleteHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewDelete[TModel any](baseModel *TModel) *Delete[TModel] {
	return &Delete[TModel]{
		whereBuilder: linq.NewWhereBuilder(baseModel),
		joinBuilder:  linq.NewJoinBuilder(),
		baseModel:    baseModel,
	}
}
