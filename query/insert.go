package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

type InsertHelper[TModel any] interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)
}

type InsertApplier interface {
	Columns() types.ExecutionColumns
}

type Insert[TModel any] struct {
	baseModel *TModel

	excludeBuilder *linq.ExcludeBuilder
}

func (h *Insert[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Insert[TModel]) Apply(applier InsertApplier) {
	h.excludeBuilder.Apply(applier)
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
