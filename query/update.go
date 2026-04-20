package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

// UpdateHelper is the per-request helper for repo.Update. It composes the
// small contracts from interfaces.go: filtering, narrowing the column set,
// and per-request RETURNING control.
type UpdateHelper[TModel any] interface {
	Filterable
	Excludable
	Returnable
}

type UpdateApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
	SetReturning(cols []types.Column)
}

type Update[TModel any] struct {
	baseModel *TModel

	excludeBuilder   *linq.ExcludeBuilder
	whereBuilder     *linq.WhereBuilder
	returningBuilder *linq.ReturningBuilder
}

func (h *Update[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Update[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *Update[TModel]) Returning(fieldsPtr ...any) {
	h.returningBuilder.Returning(fieldsPtr...)
}

func (h *Update[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Update[TModel]) Apply(applier UpdateApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	err = h.whereBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
	}
	if err := h.returningBuilder.Apply(applier); err != nil {
		return fmt.Errorf("%w: %w", ErrApplyReturningClause, err)
	}
	return nil
}

func (h *Update[TModel]) HandleFn(qFns ...func(m *TModel, h UpdateHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewUpdate[TModel any](baseModel *TModel) *Update[TModel] {
	return &Update[TModel]{
		baseModel: baseModel,

		excludeBuilder:   linq.NewExcludeBuilder(baseModel),
		whereBuilder:     linq.NewWhereBuilder(baseModel),
		returningBuilder: linq.NewReturningBuilder(baseModel),
	}
}
