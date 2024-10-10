package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/types"
)

type StringInsertBuilder struct {
	ctx     context.Context
	exclude []func(types.Column) bool
	columns []types.Column
}

func (b *StringInsertBuilder) Insert(col ...types.Column) {
	b.columns = append(b.columns, col...)
}

func (b *StringInsertBuilder) Exclude(cols ...types.Column) {
	for _, col := range cols {
		b.exclude = append(b.exclude, func(cl types.Column) bool {
			if cl == col {
				return true
			}
			return false
		})
	}
}

func (b *StringInsertBuilder) GetColumns() []types.Column {
	cols := make([]types.Column, 0, len(b.columns)-len(b.exclude))

COLUMNS:
	for _, col := range b.columns {
		if !col.IsAllowedAction(types.SQLActionInsert) {
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

func (b *StringInsertBuilder) SQL() string {
	columns := b.GetColumns()
	sql := ""
	sqlTemplate := "(%s) VALUES (%s)"
	if len(columns) < 1 {
		return sql
	}
	colsStr := ""
	valuesCount := 0
	for _, col := range columns {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		if colsStr != "" {
			colsStr += ", "
		}
		colsStr += colName
		valuesCount++
	}
	valuesSQLTemplate := strings.Repeat("?,", valuesCount)
	return fmt.Sprintf(sqlTemplate, colsStr, valuesSQLTemplate[:len(valuesSQLTemplate)-1])
}
