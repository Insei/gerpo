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
	storage types.ColumnsStorage

	vals *values
}

func NewInsert(ctx context.Context, table string, colStorage types.ColumnsStorage) *Insert {
	columns := colStorage.NewExecutionColumns(ctx, types.SQLActionInsert)
	return &Insert{
		ctx:     ctx,
		columns: columns,
		storage: colStorage,

		vals:  newValues(columns),
		table: table,
	}
}

func (i *Insert) sql() (string, error) {
	if i.table == "" {
		return "", ErrTableIsNoSet
	}
	cols := i.columns.GetAll()
	if len(cols) < 1 {
		return "", ErrEmptyColumnsInExecutionSet
	}
	sqlTemplate := "(%s) VALUES (%s)"
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
	return fmt.Sprintf("INSERT INTO %s "+sqlTemplate, i.table, colsStr, valuesSQLTemplate[:len(valuesSQLTemplate)-1]), nil
}

func (i *Insert) Columns() types.ExecutionColumns {
	return i.columns
}

func (i *Insert) ColumnsStorage() types.ColumnsStorage {
	return i.storage
}

func (i *Insert) SQL(opts ...Option) (string, []any, error) {
	for _, opt := range opts {
		opt(i.vals)
	}
	sql, err := i.sql()
	if err != nil {
		return "", nil, err
	}
	return sql, i.vals.values, nil
}
