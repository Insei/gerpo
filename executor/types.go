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
type DBAdapter extypes.DBAdapter

type Executor[TModel any] interface {
	GetOne(ctx context.Context, stmt Stmt) (*TModel, error)
	GetMultiple(ctx context.Context, stmt Stmt) ([]*TModel, error)
	InsertOne(ctx context.Context, stmt Stmt, model *TModel) error
	Update(ctx context.Context, stmt Stmt, model *TModel) (int64, error)
	Count(ctx context.Context, stmt CountStmt) (uint64, error)
	Delete(ctx context.Context, stmt CountStmt) (int64, error)
	Tx(tx Tx) (Executor[TModel], error)
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
