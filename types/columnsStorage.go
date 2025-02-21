package types

import (
	"context"
	"fmt"
	"slices"

	"github.com/insei/fmap/v3"
)

// ColumnsStorage defines an interface for managing a collection of database columns.
type ColumnsStorage interface {

	// AsSlice returns all stored columns as a slice of type Column.
	AsSlice() []Column

	// NewExecutionColumns creates a new ExecutionColumns instance for the specified SQLAction within the provided context.
	NewExecutionColumns(ctx context.Context, action SQLAction) ExecutionColumns

	// GetByFieldPtr retrieves a Column by using the provided model and field pointer, returning an error if the Column is not found.
	GetByFieldPtr(model any, fieldPtr any) (Column, error)

	// Get checks if the specified field exists and returns the corresponding Column along with a boolean indicating success.
	Get(f fmap.Field) (Column, bool)

	// Add adds a new column to the storage, incorporating it into the collection of managed columns.
	Add(column Column)
}

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

// ExecutionColumns represents an interface to manage and interact with a collection of database execution columns.
// It provides functionality to exclude columns, retrieve all columns, fetch columns by field pointers, and extract model data.
type ExecutionColumns interface {

	// Exclude removes the specified columns from the existing collection of execution columns, effectively excluding them from usage.
	Exclude(...Column)

	// GetAll retrieves and returns all the columns contained within the execution columns as a slice.
	GetAll() []Column

	// GetByFieldPtr retrieves a Column based on the provided model and field pointer.
	// The method allows fetching specific columns related to the field in the execution context.
	GetByFieldPtr(model any, fieldPtr any) Column

	// GetModelPointers retrieves a slice of pointers to the fields of the given model based on the current execution columns.
	GetModelPointers(model any) []any

	// GetModelValues retrieves the values of the model's fields mapped to the execution columns and returns them as a slice.
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
		for _, col := range cols {
			if col == column {
				return true
			}
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
	ok := false
	for _, c := range b.columns {
		if c == col {
			ok = true
			break
		}
	}
	if !ok {
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
