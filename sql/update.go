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
	cols := make([]types.Column, 0, len(b.columns)-len(b.exclude))

COLUMNS:
	for _, col := range b.columns {
		if !col.IsAllowedAction(types.SQLActionUpdate) {
			//TODO: log
			continue
		}
		for _, exclude := range b.exclude {
			if exclude(col) {
				continue COLUMNS
			}
		}
		cols = append(cols, col)
	}
	return cols
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
