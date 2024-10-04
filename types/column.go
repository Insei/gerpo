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
	SQLActionUpdate = SQLAction("update")
	SQLActionSort   = SQLAction("sort")
)

type Column interface {
	SQLFilterGetter
	IsAllowedAction(a SQLAction) bool
	ToSQL(ctx context.Context) string
	GetPtr(model any) any
	GetField() fmap.Field
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

type ColumnsStorage struct {
	m map[fmap.Field]Column
	s []Column
}

func NewColumnsStorage() *ColumnsStorage {
	return &ColumnsStorage{
		s: make([]Column, 0),
		m: make(map[fmap.Field]Column),
	}
}

func (c *ColumnsStorage) AsSlice() []Column {
	return c.s
}

func (c *ColumnsStorage) Get(f fmap.Field) (Column, bool) {
	column, ok := c.m[f]
	return column, ok
}

func (c *ColumnsStorage) Add(f fmap.Field, column Column) {
	c.m[f] = column
	c.s = append(c.s, column)
}
