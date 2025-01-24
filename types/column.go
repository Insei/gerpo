package types

import (
	"context"
	"slices"

	"github.com/insei/fmap/v3"
)

type SQLAction string

const (
	SQLActionSelect = SQLAction("select")
	SQLActionInsert = SQLAction("insert")
	SQLActionGroup  = SQLAction("group")
	SQLActionUpdate = SQLAction("update")
	SQLActionSort   = SQLAction("sort")
)

type Column interface {
	SQLFilterGetter
	IsAllowedAction(a SQLAction) bool
	GetAllowedActions() []SQLAction
	ToSQL(ctx context.Context) string
	GetPtr(model any) any
	GetField() fmap.Field
	Name() (string, bool)
	Table() (string, bool)
}

type ColumnsGetter interface {
	GetColumns() []Column
}

func NewColumnBase(field fmap.Field, toSQLFn func(ctx context.Context) string, filters SQLFilterManager) *ColumnBase {
	return &ColumnBase{
		Field:          field,
		ToSQL:          toSQLFn,
		AllowedActions: make([]SQLAction, 0),
		Filters:        filters,
		GetPtr: func(model any) any {
			return field.GetPtr(model)
		},
	}
}

type ColumnBase struct {
	Field          fmap.Field
	ToSQL          func(ctx context.Context) string
	AllowedActions []SQLAction
	Filters        SQLFilterManager
	GetPtr         func(model any) any
}

func (c *ColumnBase) IsAllowedAction(act SQLAction) bool {
	return slices.Contains(c.AllowedActions, act)
}
