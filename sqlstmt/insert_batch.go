package sqlstmt

import (
	"context"
	"strings"

	"github.com/insei/gerpo/types"
)

// InsertBatch emits a multi-row INSERT ... VALUES (...), (...), ... [RETURNING ...].
// The executor invokes SetModels with a chunk of models, then SQL() emits the SQL
// and collects values for that chunk. Chunking at the placeholder limit happens
// in the executor, not here — this stmt only renders whatever rows it was given.
type InsertBatch struct {
	ctx       context.Context
	table     string
	columns   types.ExecutionColumns
	storage   types.ColumnsStorage
	returning []types.Column
	models    []any
}

func NewInsertBatch(ctx context.Context, table string, colStorage types.ColumnsStorage) *InsertBatch {
	columns := colStorage.NewExecutionColumns(ctx, types.SQLActionInsert)
	return &InsertBatch{
		ctx:       ctx,
		table:     table,
		columns:   columns,
		storage:   colStorage,
		returning: collectReturning(colStorage, types.SQLActionInsert),
	}
}

func (b *InsertBatch) Columns() types.ExecutionColumns      { return b.columns }
func (b *InsertBatch) ColumnsStorage() types.ColumnsStorage { return b.storage }

// ReturningColumns reports the columns that should appear in RETURNING. Empty
// means the executor stays on the plain ExecContext path for this batch.
func (b *InsertBatch) ReturningColumns() []types.Column { return b.returning }

// SetReturning replaces the returning column set — used by the per-request
// query.InsertManyHelper.Returning(...) override.
func (b *InsertBatch) SetReturning(cols []types.Column) { b.returning = cols }

// SetModels gives the stmt the rows for the next SQL() call. The executor
// calls it once per chunk and then calls SQL() to emit that chunk's statement.
func (b *InsertBatch) SetModels(models []any) { b.models = models }

// SQL builds the multi-row INSERT for the models currently held by SetModels.
// Placeholder capping is the executor's concern — this function trusts its input.
func (b *InsertBatch) SQL(_ ...Option) (string, []any, error) {
	if b.table == "" {
		return "", nil, ErrTableIsNoSet
	}
	cols := b.columns.GetAll()
	if len(cols) < 1 {
		return "", nil, ErrEmptyColumnsInExecutionSet
	}
	if len(b.models) == 0 {
		return "", nil, nil
	}

	// Column names for INSERT (...) clause. Virtual columns skip (no Name()).
	var names []string
	for _, col := range cols {
		name, ok := col.Name()
		if !ok {
			continue
		}
		names = append(names, name)
	}
	valsPerRow := len(names)
	if valsPerRow == 0 {
		return "", nil, ErrEmptyColumnsInExecutionSet
	}

	rowTemplate := "(" + strings.TrimRight(strings.Repeat("?,", valsPerRow), ",") + ")"

	sb := strings.Builder{}
	sb.Grow(64 + len(names)*8 + len(b.models)*(valsPerRow*3+4))
	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(names, ", "))
	sb.WriteString(") VALUES ")

	allValues := make([]any, 0, len(b.models)*valsPerRow)
	for i, m := range b.models {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(rowTemplate)
		allValues = append(allValues, b.columns.GetModelValues(m)...)
	}
	appendReturning(&sb, b.returning)
	return sb.String(), allValues, nil
}
