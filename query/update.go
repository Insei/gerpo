package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type UpdateUserHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any)
}

type UpdateHelper[TModel any] interface {
	UpdateUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h UpdateUserHelper[TModel]))
}

type updateHelper[TModel any] struct {
	core           *linq.CoreBuilder
	excludeBuilder *linq.ExcludeBuilder
	whereBuilder   *linq.WhereBuilder
}

func (h *updateHelper[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *updateHelper[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *updateHelper[TModel]) Apply(sqlBuilder *sql.StringBuilder) {
	h.excludeBuilder.Apply(sqlBuilder.SelectBuilder())
	h.whereBuilder.Apply(sqlBuilder.WhereBuilder())
}

func (h *updateHelper[TModel]) HandleFn(qFns ...func(m *TModel, h UpdateUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func newUpdateHelper[TModel any](core *linq.CoreBuilder) *updateHelper[TModel] {
	return &updateHelper[TModel]{
		core:           core,
		excludeBuilder: linq.NewExcludeBuilder(core, types.SQLActionUpdate),
		whereBuilder:   linq.NewWhereBuilder(core),
	}
}
func NewUpdateHelper[TModel any](core *linq.CoreBuilder) UpdateHelper[TModel] {
	return newUpdateHelper[TModel](core)
}
