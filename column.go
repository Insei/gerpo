package gerpo

import (
	"fmt"
	"slices"

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

// FieldConfig is what Field(ptr) returns. It embeds *column.Builder so the regular
// column-shaping methods (WithOmitOnUpdate, WithAlias, WithColumnName, ...)
// are callable directly on the result of Field. To configure a virtual column
// instead, call AsVirtual — that swaps the registered column-builder for a
// virtual-builder and returns the latter.
type FieldConfig[TModel any] struct {
	*column.Builder
	cb    *ColumnBuilder[TModel]
	field fmap.Field
}

// AsVirtual converts the field configuration into a virtual column.
// The column-builder created by Field is removed from the registration list
// and replaced with a fresh virtual.Builder; further configuration must happen
// through the returned *virtual.Builder (Compute / Aggregate / Filter).
func (f *FieldConfig[TModel]) AsVirtual() *virtual.Builder {
	for i, b := range f.cb.builders {
		if b == f.Builder {
			f.cb.builders = slices.Delete(f.cb.builders, i, i+1)
			break
		}
	}
	vb := virtual.NewBuilder(f.field)
	f.cb.builders = append(f.cb.builders, vb)
	return vb
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

// Field registers a model field as a regular SQL column and returns a *FieldConfig
// for chained configuration. Use the embedded *column.Builder methods to refine
// the column (WithOmitOnUpdate, WithAlias, ...) or call AsVirtual() to turn
// it into a virtual column instead.
func (b *ColumnBuilder[TModel]) Field(fieldPtr any) *FieldConfig[TModel] {
	field, err := b.getFmapField(fieldPtr)
	if err != nil {
		b.errors = append(b.errors, err)
	}
	cb := column.NewBuilder(field)
	cb.WithTable(b.table)
	b.builders = append(b.builders, cb)
	return &FieldConfig[TModel]{
		Builder: cb,
		cb:      b,
		field:   field,
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
				cbCasted.OmitOnInsert().OmitOnUpdate()
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
