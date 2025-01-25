package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

type UpdateHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any)
}

type UpdateApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Columns() types.ExecutionColumns
	Where() sqlpart.Where
}
type Update[TModel any] struct {
	baseModel *TModel

	excludeBuilder *linq.ExcludeBuilder
	whereBuilder   *linq.WhereBuilder
}

func (h *Update[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *Update[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Update[TModel]) Apply(applier UpdateApplier) {
	h.excludeBuilder.Apply(applier)
	h.whereBuilder.Apply(applier)
}

func (h *Update[TModel]) HandleFn(qFns ...func(m *TModel, h UpdateHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func NewUpdate[TModel any](baseModel *TModel) *Update[TModel] {
	return &Update[TModel]{
		baseModel: baseModel,

		excludeBuilder: linq.NewExcludeBuilder(baseModel),
		whereBuilder:   linq.NewWhereBuilder(baseModel),
	}
}
