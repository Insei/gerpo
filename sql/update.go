package sql

import (
	"context"
	"slices"

	"github.com/insei/gerpo/types"
)

type StringUpdateBuilder struct {
	ctx     context.Context
	columns []types.Column
	//TODO: Add columns cache for Get columns
}

func (b *StringUpdateBuilder) Exclude(cols ...types.Column) {
	b.columns = deleteFunc(b.columns, func(column types.Column) bool {
		if slices.Contains(cols, column) {
			return true
		}
		return false
	})
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
