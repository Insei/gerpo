package sqlstmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/types"
)

type Insert struct {
	ctx context.Context

	table   string
	columns types.ExecutionColumns

	vals *values
}

func NewInsert(ctx context.Context, table string, colStorage types.ColumnsStorage) *Insert {
	columns := colStorage.NewExecutionColumns(ctx, types.SQLActionInsert)
	return &Insert{
		ctx:     ctx,
		columns: columns,

		vals:  newValues(columns),
		table: table,
	}
}

func (i *Insert) sql() string {
	cols := i.columns.GetAll()
	sql := ""
	sqlTemplate := "(%s) VALUES (%s)"
	if len(cols) < 1 {
		return sql
	}
	colsStr := ""
	valuesCount := 0
	for _, col := range cols {
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
	return fmt.Sprintf("INSERT INTO %s "+sqlTemplate, i.table, colsStr, valuesSQLTemplate[:len(valuesSQLTemplate)-1])
}

func (i *Insert) Columns() types.ExecutionColumns {
	return i.columns
}

func (i *Insert) SQL(opts ...Option) (string, []any) {
	for _, opt := range opts {
		opt(i.vals)
	}
	return i.sql(), i.vals.values
}
