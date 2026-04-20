package executor

import (
	"context"
	"fmt"

	extypes "github.com/insei/gerpo/executor/types"
	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

var ErrNoInsertedRows = fmt.Errorf("failed to insert: inserted 0 rows")
var ErrNoRows = fmt.Errorf("executor: no rows in result set")

type Tx = extypes.Tx
type ExecQuery = extypes.ExecQuery
type Adapter extypes.Adapter

type Executor[TModel any] interface {
	GetOne(ctx context.Context, stmt Stmt) (*TModel, error)
	GetMultiple(ctx context.Context, stmt Stmt) ([]*TModel, error)
	InsertOne(ctx context.Context, stmt Stmt, model *TModel) error
	InsertMany(ctx context.Context, stmt BatchStmt, models []*TModel) (int64, error)
	Update(ctx context.Context, stmt Stmt, model *TModel) (int64, error)
	Count(ctx context.Context, stmt CountStmt) (uint64, error)
	Delete(ctx context.Context, stmt CountStmt) (int64, error)
}

type CountStmt interface {
	SQL(...sqlstmt.Option) (string, []any, error)
}

type Columns interface {
	GetModelPointers(model any) []any
	GetModelValues(model any) []any
}

type Stmt interface {
	CountStmt
	Columns() types.ExecutionColumns
}

// BatchStmt is the shape the executor expects for multi-row writes. The
// executor feeds a chunk of models via SetModels, then asks for the SQL — this
// lets one InsertBatch value render many chunked statements per call.
type BatchStmt interface {
	Stmt
	SetModels(models []any)
}

// ReturningStmt is an optional capability of write statements (Insert / Update)
// that can emit a RETURNING clause. The returned slice lists the columns
// scanned back into the caller's model after the SQL runs; an empty slice
// disables the RETURNING path so the executor stays on ExecContext.
type ReturningStmt interface {
	ReturningColumns() []types.Column
}
