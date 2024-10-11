package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type DeleteUserHelper[TModel any] interface {
	Where() types.WhereTarget
}

type DeleteHelper[TModel any] interface {
	DeleteUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h DeleteUserHelper[TModel]))
}

type deleteHelper[TModel any] struct {
	core         *linq.CoreBuilder
	whereBuilder *linq.WhereBuilder
}

func (h *deleteHelper[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *deleteHelper[TModel]) Apply(sqlBuilder *sql.StringBuilder) {
	h.whereBuilder.Apply(sqlBuilder.WhereBuilder())
}

func (h *deleteHelper[TModel]) HandleFn(qFns ...func(m *TModel, h DeleteUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func newDeleteHelper[TModel any](core *linq.CoreBuilder) *deleteHelper[TModel] {
	return &deleteHelper[TModel]{
		whereBuilder: linq.NewWhereBuilder(core),
		core:         core,
	}
}

func NewDeleteHelper[TModel any](core *linq.CoreBuilder) DeleteHelper[TModel] {
	return newDeleteHelper[TModel](core)
}
