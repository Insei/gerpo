package gerpo

import (
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/column"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

type columnBuild interface {
	Build() types.Column
}

type ColumnBuilder[TModel any] struct {
	table         string
	model         *TModel
	columns       types.ColumnsStorage
	fieldsStorage fmap.Storage
	builders      []columnBuild
}

type ColumnTypeSelector[TModel any] struct {
	cb       *ColumnBuilder[TModel]
	fieldPtr any
}

// AsVirtual creates a new virtual column builder for the specified field and appends it to the list of column builders.
func (s ColumnTypeSelector[TModel]) AsVirtual() *virtual.Builder {
	field := s.cb.getFmapField(s.fieldPtr)
	vb := virtual.NewBuilder(field)
	s.cb.builders = append(s.cb.builders, vb)
	return vb
}

// AsColumn initializes a column builder for the specified field and appends it to the column builder’s list of builders.
func (s ColumnTypeSelector[TModel]) AsColumn() *column.Builder {
	field := s.cb.getFmapField(s.fieldPtr)
	b := column.NewBuilder(field)
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

func (b *ColumnBuilder[TModel]) getFmapField(fieldPtr any) fmap.Field {
	field, err := b.fieldsStorage.GetFieldByPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	return field
}

// Field initializes the building process for a specific field of the model and returns a ColumnTypeSelector for further configuration.
func (b *ColumnBuilder[TModel]) Field(fieldPtr any) *ColumnTypeSelector[TModel] {
	return &ColumnTypeSelector[TModel]{
		cb:       b,
		fieldPtr: fieldPtr,
	}
}

func (b *ColumnBuilder[TModel]) build() types.ColumnsStorage {
	for _, cb := range b.builders {
		cl := cb.Build()
		// Makes column
		if table, ok := cl.Table(); !ok || table == "" || table != b.table {
			if cbCasted, ok := cb.(*column.Builder); ok {
				cbCasted.WithInsertProtection().WithUpdateProtection()
				cl = cbCasted.Build()
			}
		}
		b.columns.Add(cl)
	}
	return b.columns
}
