package sql

import (
	"context"

	"github.com/insei/gerpo/types"
)

type StringUpdateBuilder struct {
	ctx     context.Context
	exclude []func(types.Column) bool
	columns []types.Column
	//TODO: Add columns cache for Get columns
}

func (b *StringUpdateBuilder) Columns(col ...types.Column) {
	b.columns = append(b.columns, col...)
}

func (b *StringUpdateBuilder) GetColumns() []types.Column {
	return b.columns
}

func (b *StringUpdateBuilder) SQL() string {
	columns := b.GetColumns()
	colsStr := ""
	if len(columns) < 1 {
		return colsStr
	}
	for _, col := range columns {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		colsStr += colName + " = ?, "
	}
	return colsStr[:len(colsStr)-2]
}
