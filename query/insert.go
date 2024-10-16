package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
)

type InsertUserHelper[TModel any] interface {
	Exclude(fieldsPtr ...any)
}

type InsertHelper[TModel any] interface {
	InsertUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h InsertUserHelper[TModel]))
}

type insertHelper[TModel any] struct {
	core           *linq.CoreBuilder
	excludeBuilder *linq.ExcludeBuilder
}

func (h *insertHelper[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *insertHelper[TModel]) Apply(sqlBuilder *sql.StringBuilder) {
	h.excludeBuilder.Apply(sqlBuilder.InsertBuilder())
}

func (h *insertHelper[TModel]) HandleFn(qFns ...func(m *TModel, h InsertUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func newInsertHelper[TModel any](core *linq.CoreBuilder) *insertHelper[TModel] {
	return &insertHelper[TModel]{
		excludeBuilder: linq.NewExcludeBuilder(core),
		core:           core,
	}
}

func NewInsertHelper[TModel any](core *linq.CoreBuilder) InsertHelper[TModel] {
	return newInsertHelper[TModel](core)
}
