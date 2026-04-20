package sqlstmt

import (
	"context"
	"strings"

	"github.com/insei/gerpo/types"
)

type Insert struct {
	ctx context.Context

	table     string
	columns   types.ExecutionColumns
	storage   types.ColumnsStorage
	returning []types.Column

	vals *values
}

func NewInsert(ctx context.Context, table string, colStorage types.ColumnsStorage) *Insert {
	columns := colStorage.NewExecutionColumns(ctx, types.SQLActionInsert)
	return &Insert{
		ctx:       ctx,
		columns:   columns,
		storage:   colStorage,
		returning: collectReturning(colStorage, types.SQLActionInsert),

		vals:  newValues(columns),
		table: table,
	}
}

func (i *Insert) Columns() types.ExecutionColumns {
	return i.columns
}

func (i *Insert) ColumnsStorage() types.ColumnsStorage {
	return i.storage
}

// ReturningColumns reports the columns that should be scanned back from a
// RETURNING clause. Empty slice means the executor takes the plain ExecContext
// path (no rows returned).
func (i *Insert) ReturningColumns() []types.Column {
	return i.returning
}

// SetReturning replaces the returning column set — used by the per-request
// query.InsertHelper.Returning(...) override. Pass nil/empty slice to disable
// RETURNING for the call.
func (i *Insert) SetReturning(cols []types.Column) {
	i.returning = cols
}

func (i *Insert) SQL(opts ...Option) (string, []any, error) {
	for _, opt := range opts {
		opt(i.vals)
	}
	if i.table == "" {
		return "", nil, ErrTableIsNoSet
	}
	cols := i.columns.GetAll()
	if len(cols) < 1 {
		return "", nil, ErrEmptyColumnsInExecutionSet
	}
	sb := strings.Builder{}
	sb.Grow(128)
	sb.WriteString("INSERT INTO ")
	sb.WriteString(i.table)
	sb.WriteString(" (")
	lenAtStart := sb.Len()
	valuesCount := 0
	for _, col := range cols {
		colName, ok := col.Name()
		if !ok {
			continue
		}
		if sb.Len() > lenAtStart {
			sb.WriteString(", ")
		}
		sb.WriteString(colName)
		valuesCount++
	}
	sb.WriteString(") VALUES (")
	valuesSQLTemplate := strings.Repeat("?,", valuesCount)
	sb.WriteString(valuesSQLTemplate[:len(valuesSQLTemplate)-1] + ")")
	appendReturning(&sb, i.returning)
	return sb.String(), i.vals.values, nil
}
