package query

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// DeleteHelper is the per-request helper for repo.Delete. It only filters —
// see interfaces.go for the Filterable contract.
type DeleteHelper[TModel any] interface {
	Filterable
}

type DeleteApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Where() sqlpart.Where
	Join() sqlpart.Join
	Ctx() context.Context
}

type Delete[TModel any] struct {
	baseModel *TModel

	whereBuilder *linq.WhereBuilder
	joinBuilder  *linq.JoinBuilder
}

func (h *Delete[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Delete[TModel]) Apply(applier DeleteApplier) error {
	err := h.whereBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
	}
	err = h.joinBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyJoinClause, err)
	}
	return nil
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
