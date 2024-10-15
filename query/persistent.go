package query

import (
	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type PersistentUserHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any)
	GroupBy(fieldsPtr ...any) PersistentUserHelper[TModel]
}

type PersistentHelper[TModel any] interface {
	PersistentUserHelper[TModel]
	SQLApply
	HandleFn(qFns ...func(m *TModel, h PersistentUserHelper[TModel]))
}

type persistentHelper[TModel any] struct {
	core           *linq.CoreBuilder
	excludeBuilder *linq.ExcludeBuilder
	whereBuilder   *linq.WhereBuilder
	groupBuilder   *linq.GroupBuilder
}

func (h *persistentHelper[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *persistentHelper[TModel]) Exclude(fieldsPtr ...any) {
	h.excludeBuilder.Exclude(fieldsPtr...)
}

func (h *persistentHelper[TModel]) GroupBy(fieldsPtr ...any) PersistentUserHelper[TModel] {
	h.groupBuilder.GroupBy(fieldsPtr...)
	return h
}

func (h *persistentHelper[TModel]) HandleFn(qFns ...func(m *TModel, h PersistentUserHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.core.Model().(*TModel), h)
	}
}

func (h *persistentHelper[TModel]) Apply(sqlBuilder *sql.StringBuilder) {
	h.excludeBuilder.Apply(sqlBuilder.SelectBuilder())
	h.whereBuilder.Apply(sqlBuilder.WhereBuilder())
	h.groupBuilder.Apply(sqlBuilder.GroupBuilder())
}

func newPersistentHelper[TModel any](core *linq.CoreBuilder, action types.SQLAction) *persistentHelper[TModel] {
	return &persistentHelper[TModel]{
		core:           core,
		excludeBuilder: linq.NewExcludeBuilder(core, action),
		whereBuilder:   linq.NewWhereBuilder(core),
		groupBuilder:   linq.NewGroupBuilder(core),
	}
}

func NewPersistentHelper[TModel any](core *linq.CoreBuilder, action types.SQLAction) PersistentHelper[TModel] {
	return newPersistentHelper[TModel](core, action)
}
