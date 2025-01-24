package executor

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/insei/gerpo/sqlstmt"
	"github.com/insei/gerpo/types"
)

var ErrNoInsertedRows = fmt.Errorf("failed to insert: inserted 0 rows")
var ErrTxDBNotSame = fmt.Errorf("tx db not the same as in main executor")

type ExecQuery interface {
	ExecContext(ctx context.Context, query string, args ...any) (dbsql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*dbsql.Rows, error)
}

type Executor[TModel any] interface {
	GetOne(ctx context.Context, stmt Stmt) (*TModel, error)
	GetMultiple(ctx context.Context, stmt Stmt) ([]*TModel, error)
	InsertOne(ctx context.Context, stmt Stmt, model *TModel) error
	Update(ctx context.Context, stmt Stmt, model *TModel) (int64, error)
	Count(ctx context.Context, stmt CountStmt) (uint64, error)
	Delete(ctx context.Context, stmt CountStmt) (int64, error)
	Tx(tx *Tx) (Executor[TModel], error)
}

type CountStmt interface {
	SQL(...sqlstmt.Option) (string, []any)
}

type Columns interface {
	GetModelPointers(model any) []any
	GetModelValues(model any) []any
}

type Stmt interface {
	CountStmt
	Columns() types.ExecutionColumns
}
