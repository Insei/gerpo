package query

import (
	"context"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

// PersistentHelper is an interface for building and executing queries with conditions, joins, and group operations.
type PersistentHelper[TModel any] interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any) PersistentHelper[TModel]
	// Where defines the starting point for building conditions in a query, returning a types.WhereTarget interface.
	Where() types.WhereTarget

	// GroupBy groups the query results by the specified fields, accepting variadic pointers to fields for grouping operations.
	GroupBy(fieldsPtr ...any) PersistentHelper[TModel]

	// LeftJoin adds a LEFT JOIN clause to the query using a provided function that returns the SQL join statement.
	LeftJoin(func(ctx context.Context) string) PersistentHelper[TModel]

	// InnerJoin adds a INNER JOIN clause to the query using a provided function that returns the SQL join statement.
	InnerJoin(fn func(ctx context.Context) string) PersistentHelper[TModel]
}

type Persistent[TModel any] struct {
	baseModel *TModel

	excludeBuilder *linq.ExcludeBuilder
	whereBuilder   *linq.WhereBuilder
	groupBuilder   *linq.GroupBuilder
	joinBuilder    *linq.JoinBuilder
}

func (h *Persistent[TModel]) Where() types.WhereTarget {
	return h.whereBuilder
}

func (h *Persistent[TModel]) LeftJoin(fn func(ctx context.Context) string) PersistentHelper[TModel] {
	h.joinBuilder.LeftJoin(fn)
	return h
}

func (h *Persistent[TModel]) InnerJoin(fn func(ctx context.Context) string) PersistentHelper[TModel] {
	h.joinBuilder.InnerJoin(fn)
	return h
}

func (h *Persistent[TModel]) Exclude(fieldsPtr ...any) PersistentHelper[TModel] {
	h.excludeBuilder.Exclude(fieldsPtr...)
	return h
}

func (h *Persistent[TModel]) GroupBy(fieldsPtr ...any) PersistentHelper[TModel] {
	h.groupBuilder.GroupBy(fieldsPtr...)
	return h
}

func (h *Persistent[TModel]) HandleFn(qFns ...func(m *TModel, h PersistentHelper[TModel])) {
	for _, fn := range qFns {
		fn(h.baseModel, h)
	}
}

func (h *Persistent[TModel]) Apply(applier any) {
	if applier == nil {
		return
	}
	if whereApplier, ok := applier.(linq.WhereApplier); ok {
		h.whereBuilder.Apply(whereApplier)
	}

	if joinApplier, ok := applier.(linq.JoinApplier); ok {
		h.joinBuilder.Apply(joinApplier)
	}

	if excludeApplier, ok := applier.(linq.ExcludeApplier); ok {
		h.excludeBuilder.Apply(excludeApplier)
	}

	if groupApplier, ok := applier.(linq.GroupApplier); ok {
		h.groupBuilder.Apply(groupApplier)
	}
	//if sqlStr := sqlBuilder.WhereBuilder().SQL(); sqlStr != "" && !h.whereBuilder.IsEmpty() {
	//	sqlBuilder.WhereBuilder().AND()
	//}
}

func NewPersistent[TModel any](baseModel *TModel) *Persistent[TModel] {
	return &Persistent[TModel]{
		baseModel: baseModel,

		excludeBuilder: linq.NewExcludeBuilder(baseModel),
		whereBuilder:   linq.NewWhereBuilder(baseModel),
		groupBuilder:   linq.NewGroupBuilder(baseModel),
		joinBuilder:    linq.NewJoinBuilder(),
	}
}
