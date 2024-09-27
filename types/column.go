package types

import (
	"context"
	"slices"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/filter"
)

type IQueryBuilder interface {
}

type AllowedAction string

const (
	ActionRead   = AllowedAction("read")
	ActionUpdate = AllowedAction("update")
	ActionSort   = AllowedAction("sort")
)

type Columns map[fmap.Field]Column
type Column interface {
	filter.SQLFilterGetter
	IsAllowedAction(a AllowedAction) bool
	ToSQL(ctx context.Context) string
	GetPtr(model any) any
}

func NewColumnBase(field fmap.Field, toSQLFn func(ctx context.Context) string) *ColumnBase {
	return &ColumnBase{
		ToSQL:          toSQLFn,
		AllowedActions: make([]AllowedAction, 0),
		Filters:        filter.NewForField(field),
		GetPtr: func(model any) any {
			return field.GetPtr(model)
		},
	}
}

type ColumnBase struct {
	ToSQL          func(ctx context.Context) string
	AllowedActions []AllowedAction
	Filters        filter.SQLFilterManager
	GetPtr         func(model any) any
}

func (c *ColumnBase) IsAllowedAction(act AllowedAction) bool {
	return slices.Contains(c.AllowedActions, act)
}
