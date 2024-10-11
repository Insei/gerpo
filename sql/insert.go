package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/types"
)

type StringInsertBuilder struct {
	ctx     context.Context
	columns []types.Column
}

func (b *StringInsertBuilder) Columns(col ...types.Column) {
	for _, col := range col {
		if !col.IsAllowedAction(types.SQLActionInsert) {
			//TODO: log
			continue
		}
		b.columns = append(b.columns, col)
	}
}

func (b *StringInsertBuilder) GetColumns() []types.Column {
	return b.columns
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
