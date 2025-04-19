package query

import (
	"fmt"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

type InsertHelper[TModel any] interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)
	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(fieldsPtr ...any)
}

type InsertApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
}

type Insert[TModel any] struct {
	baseModel *TModel

	excludeBuilder *linq.ExcludeBuilder
}

func (h *Insert[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Insert[TModel]) Only(fieldPointers ...any) {
	h.excludeBuilder.Only(fieldPointers...)
}

func (h *Insert[TModel]) Apply(applier InsertApplier) error {
	err := h.excludeBuilder.Apply(applier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
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

		excludeBuilder: linq.NewExcludeBuilder(baseModel),
	}
}
