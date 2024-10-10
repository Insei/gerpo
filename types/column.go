package types

import (
	"context"
	"fmt"
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
	m       map[fmap.Field]Column
	s       []Column
	storage fmap.Storage
	model   any
}

func NewEmptyColumnsStorage(fields fmap.Storage) *ColumnsStorage {
	return &ColumnsStorage{
		s:       make([]Column, 0),
		m:       make(map[fmap.Field]Column),
		storage: fields,
	}
}

func (c *ColumnsStorage) AsSlice() []Column {
	return c.s
}

func (c *ColumnsStorage) GetByFieldPtr(model any, fieldPtr any) (Column, error) {
	field, err := c.storage.GetFieldByPtr(model, fieldPtr)
	if err != nil {
		return nil, err
	}
	column, ok := c.m[field]
	if !ok {
		return nil, fmt.Errorf("column for field %s was not found in configured columns", field.GetStructPath())
	}
	return column, nil
}

func (c *ColumnsStorage) Get(f fmap.Field) (Column, bool) {
	column, ok := c.m[f]
	return column, ok
}

func (c *ColumnsStorage) Add(f fmap.Field, column Column) {
	c.m[f] = column
	c.s = append(c.s, column)
}
