package executor

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/insei/gerpo/sql"
)

var ErrNoInsertedRows = fmt.Errorf("failed to insert: inserted 0 rows")
var ErrTxDBNotSame = fmt.Errorf("tx db not the same as in main executor")

type ExecQuery interface {
	ExecContext(ctx context.Context, query string, args ...any) (dbsql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*dbsql.Rows, error)
}

type Executor[TModel any] interface {
	GetOne(ctx context.Context, selectStmt sql.StmtSelect) (*TModel, error)
	GetMultiple(ctx context.Context, selectStmt sql.StmtSelect) ([]*TModel, error)
	InsertOne(ctx context.Context, model *TModel, stmtModel sql.StmtModel) error
	Update(ctx context.Context, model *TModel, stmtModel sql.StmtModel) (int64, error)
	Count(ctx context.Context, stmt sql.Stmt) (uint64, error)
	Delete(ctx context.Context, stmt sql.Stmt) (int64, error)
	Tx(tx *Tx) (Executor[TModel], error)
}
