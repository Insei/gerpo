package types

import (
	"context"
	"fmt"
	"slices"

	"github.com/insei/fmap/v3"
)

type ColumnsStorage interface {
	AsSlice() []Column
	NewExecutionColumns(ctx context.Context, action SQLAction) ExecutionColumns
	GetByFieldPtr(model any, fieldPtr any) (Column, error)
	Get(f fmap.Field) (Column, bool)
	Add(column Column)
}

type columnsStorage struct {
	m       map[fmap.Field]Column
	s       []Column
	act     map[SQLAction][]Column
	storage fmap.Storage
	model   any
}

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
	return &executionColumns{
		storage: c,
		columns: cols,
		ctx:     ctx,
	}
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

type ExecutionColumns interface {
	Exclude(...Column)
	GetAll() []Column
	GetByFieldPtr(model any, fieldPtr any) Column
	GetModelPointers(model any) []any
	GetModelValues(model any) []any
}

type executionColumns struct {
	storage *columnsStorage
	ctx     context.Context
	columns []Column
}

// deleteFunc Modified deleteFunc from slices packages without clean element,
// removes any elements from s for which del returns true,
// returning the modified slice.
// deleteFunc zeroes the elements between the new length and the original length.
func deleteFunc[S ~[]E, E any](s S, del func(E) bool) S {
	i := slices.IndexFunc(s, del)
	if i == -1 {
		return s
	}

	var newSlice []E = make([]E, 0, len(s))

	for j := 0; j < len(s); j++ {
		if v := s[j]; !del(s[j]) {
			newSlice = append(newSlice, v)
		}
	}

	return newSlice
}

func (b *executionColumns) Exclude(cols ...Column) {
	b.columns = deleteFunc(b.columns, func(column Column) bool {
		if slices.Contains(cols, column) {
			return true
		}
		return false
	})
}

func (b *executionColumns) GetAll() []Column {
	return b.columns
}

func (b *executionColumns) GetByFieldPtr(model any, fieldPtr any) Column {
	col, err := b.storage.GetByFieldPtr(model, fieldPtr)
	if err != nil {
		panic(err)
	}
	if slices.Contains(b.columns, col) {
		panic("trying to get excluded column?")
	}
	return col
}

// GetModelPointers returns a slice of pointers to the fields of the provided model corresponding to the execution columns.
func (b *executionColumns) GetModelPointers(model any) []any {
	pointers := make([]any, 0, len(b.columns))
	for _, col := range b.columns {
		pointers = append(pointers, col.GetPtr(model))
	}
	return pointers
}

// GetModelValues retrieves the values of the model's fields associated with the execution columns and returns them as a slice.
func (b *executionColumns) GetModelValues(model any) []any {
	values := make([]any, 0, len(b.columns))
	for _, col := range b.columns {
		values = append(values, col.GetField().Get(model))
	}
	return values
}
