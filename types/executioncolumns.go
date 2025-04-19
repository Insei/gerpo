package types

import (
	"context"
	"fmt"
	"slices"
)

type executionColumns struct {
	storage *columnsStorage
	ctx     context.Context
	columns []Column
}

func newExecutionColumns(ctx context.Context, storage *columnsStorage, columns []Column) *executionColumns {
	return &executionColumns{
		storage: storage,
		ctx:     ctx,
		columns: columns,
	}
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

func (b *executionColumns) Only(cols ...Column) {
	b.columns = cols
}

func (b *executionColumns) GetAll() []Column {
	return b.columns
}

func (b *executionColumns) GetByFieldPtr(model any, fieldPtr any) (Column, error) {
	col, err := b.storage.GetByFieldPtr(model, fieldPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get column by field ptr: %w", err)
	}
	ok := false
	for _, c := range b.columns {
		if c == col {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("column %s not found in execution columns", col.GetField().GetStructPath())
	}
	return col, nil
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
