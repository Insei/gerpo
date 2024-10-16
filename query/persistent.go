package query

import (
	"context"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

type PersistentUserHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any) PersistentUserHelper[TModel]
	GroupBy(fieldsPtr ...any) PersistentUserHelper[TModel]
	LeftJoin(func(ctx context.Context) string) PersistentUserHelper[TModel]
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
	joinBuilder    *linq.JoinBuilder
}

func (h *persistentHelper[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *persistentHelper[TModel]) LeftJoin(fn func(ctx context.Context) string) PersistentUserHelper[TModel] {
	h.joinBuilder.LeftJoin(fn)
	return h
}

func (h *persistentHelper[TModel]) Exclude(fieldsPtr ...any) PersistentUserHelper[TModel] {
	h.excludeBuilder.Exclude(fieldsPtr...)
	return h
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
	if sqlStr := sqlBuilder.WhereBuilder().SQL(); sqlStr != "" && !h.whereBuilder.IsEmpty() {
		sqlBuilder.WhereBuilder().AND()
	}
	h.joinBuilder.Apply(sqlBuilder.JoinBuilder())
	h.whereBuilder.Apply(sqlBuilder.WhereBuilder())
	h.excludeBuilder.Apply(sqlBuilder.SelectBuilder())
	h.excludeBuilder.Apply(sqlBuilder.UpdateBuilder())
	h.excludeBuilder.Apply(sqlBuilder.InsertBuilder())
	h.groupBuilder.Apply(sqlBuilder.GroupBuilder())
}

func newPersistentHelper[TModel any](core *linq.CoreBuilder) *persistentHelper[TModel] {
	return &persistentHelper[TModel]{
		core:           core,
		excludeBuilder: linq.NewExcludeBuilder(core),
		whereBuilder:   linq.NewWhereBuilder(core),
		groupBuilder:   linq.NewGroupBuilder(core),
		joinBuilder:    linq.NewJoinBuilder(core),
	}
}

func NewPersistentHelper[TModel any](core *linq.CoreBuilder) PersistentHelper[TModel] {
	return newPersistentHelper[TModel](core)
}
