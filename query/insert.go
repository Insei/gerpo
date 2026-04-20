package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

// InsertHelper is the per-request helper for repo.Insert. It narrows the
// column set (Excludable) and lets the caller override the RETURNING clause
// (Returnable) — see interfaces.go for the small contracts.
type InsertHelper[TModel any] interface {
	Excludable
	Returnable
}

type InsertApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	SetReturning(cols []types.Column)
}

type Insert[TModel any] struct {
	baseModel *TModel

	excludeBuilder   *linq.ExcludeBuilder
	returningBuilder *linq.ReturningBuilder
}

func (h *Insert[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Insert[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *Insert[TModel]) Returning(fieldsPtr ...any) {
	h.returningBuilder.Returning(fieldsPtr...)
}

func (h *Insert[TModel]) Apply(applier InsertApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	if err := h.returningBuilder.Apply(applier); err != nil {
		return fmt.Errorf("%w: %w", ErrApplyReturningClause, err)
	}
	return nil
}

func (h *Insert[TModel]) HandleFn(qFns ...func(m *TModel, h InsertHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewInsert[TModel any](baseModel *TModel) *Insert[TModel] {
	return &Insert[TModel]{
		baseModel: baseModel,

		excludeBuilder:   linq.NewExcludeBuilder(baseModel),
		returningBuilder: linq.NewReturningBuilder(baseModel),
	}
}
