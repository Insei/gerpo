package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

// InsertManyHelper is the per-request helper for repo.InsertMany. It mirrors
// the single-row InsertHelper: Excludable narrows the column set, Returnable
// overrides the RETURNING clause. The RETURNING set is identical for every row
// in the batch — there is no per-row override.
type InsertManyHelper[TModel any] interface {
	Excludable
	Returnable
}

type InsertManyApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	SetReturning(cols []types.Column)
}

type InsertMany[TModel any] struct {
	baseModel *TModel

	excludeBuilder   *linq.ExcludeBuilder
	returningBuilder *linq.ReturningBuilder
}

func (h *InsertMany[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *InsertMany[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *InsertMany[TModel]) Returning(fieldsPtr ...any) {
	h.returningBuilder.Returning(fieldsPtr...)
}

func (h *InsertMany[TModel]) Apply(applier InsertManyApplier) error {
	if err := h.excludeBuilder.Apply(applier); err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
	}
	if err := h.returningBuilder.Apply(applier); err != nil {
		return fmt.Errorf("%w: %w", ErrApplyReturningClause, err)
	}
	return nil
}

func (h *InsertMany[TModel]) HandleFn(qFns ...func(m *TModel, h InsertManyHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewInsertMany[TModel any](baseModel *TModel) *InsertMany[TModel] {
	return &InsertMany[TModel]{
		baseModel: baseModel,

		excludeBuilder:   linq.NewExcludeBuilder(baseModel),
		returningBuilder: linq.NewReturningBuilder(baseModel),
	}
}
