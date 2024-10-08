package gerpo

import (
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/column"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

type ColumnBuilder[TModel any] struct {
	table         string
	model         *TModel
	columns       *types.ColumnsStorage
	fieldsStorage fmap.Storage
	fields        []fmap.Field
	builders      []func() types.Column
}

func NewColumnBuilder[TModel any](table string, model *TModel, fields fmap.Storage) *ColumnBuilder[TModel] {
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

func (b *ColumnBuilder[TModel]) Column(fieldPtr any) *column.Builder {
	field := b.getFmapField(fieldPtr)
	builder := column.NewBuilder(field)
	builder.WithTable(b.table)
	b.builders = append(b.builders, builder.Build)
	b.fields = append(b.fields, field)
	return builder
}

func (b *ColumnBuilder[TModel]) Virtual(fieldPtr any) *virtual.Builder {
	field := b.getFmapField(fieldPtr)
	builder := virtual.NewBuilder(field)
	b.builders = append(b.builders, builder.Build)
	b.fields = append(b.fields, field)
	return builder
}

func (b *ColumnBuilder[TModel]) build() *types.ColumnsStorage {
	for i, buildFn := range b.builders {
		field := b.fields[i]
		cl := buildFn()
		b.columns.Add(field, cl)
	}
	return b.columns
}
