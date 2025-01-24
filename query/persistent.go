package query

import (
	"context"

	"github.com/insei/gerpo/query/linq"
	"github.com/insei/gerpo/types"
)

type PersistentHelper[TModel any] interface {
	Where() types.WhereTarget
	Exclude(fieldsPtr ...any) PersistentHelper[TModel]
	GroupBy(fieldsPtr ...any) PersistentHelper[TModel]
	LeftJoin(func(ctx context.Context) string) PersistentHelper[TModel]
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
