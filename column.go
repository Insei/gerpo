package gerpo

import (
	"fmt"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/column"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

type columnBuild interface {
	Build() (types.Column, error)
}

type ColumnBuilder[TModel any] struct {
	table         string
	model         *TModel
	columns       types.ColumnsStorage
	fieldsStorage fmap.Storage
	builders      []columnBuild
	errors        []error
}

type ColumnTypeSelector[TModel any] struct {
	cb       *ColumnBuilder[TModel]
	fieldPtr any
}

// AsVirtual creates a new virtual column builder for the specified field and appends it to the list of column builders.
func (s ColumnTypeSelector[TModel]) AsVirtual() *virtual.Builder {
	field, err := s.cb.getFmapField(s.fieldPtr)
	if err != nil {
		s.cb.errors = append(s.cb.errors, err)
	}
	vb := virtual.NewBuilder(field)
	s.cb.builders = append(s.cb.builders, vb)
	return vb
}

// AsColumn initializes a column builder for the specified field and appends it to the column builder’s list of builders.
func (s ColumnTypeSelector[TModel]) AsColumn() *column.Builder {
	field, err := s.cb.getFmapField(s.fieldPtr)
	if err != nil {
		s.cb.errors = append(s.cb.errors, err)
	}
	b := column.NewBuilder(field)
	b.WithTable(s.cb.table)
	s.cb.builders = append(s.cb.builders, b)
	return b
}

func newColumnBuilder[TModel any](table string, model *TModel, fields fmap.Storage) *ColumnBuilder[TModel] {
	return &ColumnBuilder[TModel]{
		table:         table,
		model:         model,
		columns:       types.NewEmptyColumnsStorage(fields),
		fieldsStorage: fields,
	}
}

func (b *ColumnBuilder[TModel]) getFmapField(fieldPtr any) (fmap.Field, error) {
	field, err := b.fieldsStorage.GetFieldByPtr(b.model, fieldPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get field by field pointer: %w", err)
	}
	return field, nil
}

// Field initializes the building process for a specific field of the model and returns a ColumnTypeSelector for further configuration.
func (b *ColumnBuilder[TModel]) Field(fieldPtr any) *ColumnTypeSelector[TModel] {
	return &ColumnTypeSelector[TModel]{
		cb:       b,
		fieldPtr: fieldPtr,
	}
}

func (b *ColumnBuilder[TModel]) build() (types.ColumnsStorage, error) {
	if len(b.errors) > 0 {
		return nil, b.errors[0]
	}
	for _, cb := range b.builders {
		cl, err := cb.Build()
		if err != nil {
			return nil, err
		}
		// Makes virtual/joined columns read only
		if table, ok := cl.Table(); !ok || table == "" || table != b.table {
			if cbCasted, ok := cb.(*column.Builder); ok {
				cbCasted.WithInsertProtection().WithUpdateProtection()
				cl, err = cbCasted.Build()
				if err != nil {
					return nil, fmt.Errorf("failed to build column from builder: %w", err)
				}
			}
		}
		b.columns.Add(cl)
	}
	return b.columns, nil
}
