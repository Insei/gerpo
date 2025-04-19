package types

import (
	"context"
	"fmt"

	"github.com/insei/fmap/v3"
)

type columnsStorage struct {
	m       map[fmap.Field]Column
	s       []Column
	act     map[SQLAction][]Column
	storage fmap.Storage
	model   any
}

// NewEmptyColumnsStorage creates a new empty ColumnsStorage instance with initialized internal structures.
func NewEmptyColumnsStorage(fields fmap.Storage) ColumnsStorage {
	return &columnsStorage{
		s:       make([]Column, 0),
		m:       make(map[fmap.Field]Column),
		act:     make(map[SQLAction][]Column),
		storage: fields,
	}
}

func (c *columnsStorage) AsSlice() []Column {
	return c.s
}

func (c *columnsStorage) NewExecutionColumns(ctx context.Context, action SQLAction) ExecutionColumns {
	cols, ok := c.act[action]
	if !ok {
		return nil
	}
	return newExecutionColumns(ctx, c, cols)
}

func (c *columnsStorage) GetByFieldPtr(model any, fieldPtr any) (Column, error) {
	field, err := c.storage.GetFieldByPtr(model, fieldPtr)
	if err != nil {
		return nil, err
	}
	column, ok := c.m[field]
	if !ok {
		return nil, fmt.Errorf("column for field %s was not found in configured executionColumns", field.GetStructPath())
	}
	return column, nil
}

func (c *columnsStorage) Get(f fmap.Field) (Column, bool) {
	column, ok := c.m[f]
	return column, ok
}

func (c *columnsStorage) Add(column Column) {
	c.m[column.GetField()] = column
	c.s = append(c.s, column)
	if column.IsAllowedAction(SQLActionInsert) {
		c.act[SQLActionInsert] = append(c.act[SQLActionInsert], column)
	}
	if column.IsAllowedAction(SQLActionSelect) {
		c.act[SQLActionSelect] = append(c.act[SQLActionSelect], column)
	}
	if column.IsAllowedAction(SQLActionUpdate) {
		c.act[SQLActionUpdate] = append(c.act[SQLActionUpdate], column)
	}
	if column.IsAllowedAction(SQLActionGroup) {
		c.act[SQLActionGroup] = append(c.act[SQLActionGroup], column)
	}
	if column.IsAllowedAction(SQLActionSort) {
		c.act[SQLActionSort] = append(c.act[SQLActionSort], column)
	}
}
