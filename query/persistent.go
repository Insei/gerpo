package query

import (
	"fmt"

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

	// LeftJoinOn adds a LEFT JOIN with a fixed table reference and an ON
	// clause. The SQL template stays frozen at registration time, but `?`
	// placeholders inside it can be materialized per request by an optional
	// resolver that receives the request ctx and returns a value list.
	// Omit the resolver for a static JOIN. Passing more than one resolver
	// panics. Resolver errors abort the query before it hits the driver.
	LeftJoinOn(table, on string, resolver ...linq.JoinArgsResolver) PersistentHelper[TModel]

	// InnerJoinOn is the INNER counterpart of LeftJoinOn with identical
	// resolver semantics.
	InnerJoinOn(table, on string, resolver ...linq.JoinArgsResolver) PersistentHelper[TModel]
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

func (h *Persistent[TModel]) LeftJoinOn(table, on string, resolver ...linq.JoinArgsResolver) PersistentHelper[TModel] {
	h.joinBuilder.LeftJoinOn(table, on, resolver...)
	return h
}

func (h *Persistent[TModel]) InnerJoinOn(table, on string, resolver ...linq.JoinArgsResolver) PersistentHelper[TModel] {
	h.joinBuilder.InnerJoinOn(table, on, resolver...)
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

func (h *Persistent[TModel]) Apply(applier any) error {
	if applier == nil {
		return fmt.Errorf("applier is nil")
	}
	if whereApplier, ok := applier.(linq.WhereApplier); ok {
		err := h.whereBuilder.Apply(whereApplier)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrApplyWhereClause, err)
		}
	}

	if joinApplier, ok := applier.(linq.JoinApplier); ok {
		err := h.joinBuilder.Apply(joinApplier)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrApplyJoinClause, err)
		}
	}

	if excludeApplier, ok := applier.(linq.ExcludeApplier); ok {
		err := h.excludeBuilder.Apply(excludeApplier)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrApplyExcludeColumnRules, err)
		}
	}

	if groupApplier, ok := applier.(linq.GroupApplier); ok {
		err := h.groupBuilder.Apply(groupApplier)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrApplyGroupByClause, err)
		}
	}
	return nil
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
