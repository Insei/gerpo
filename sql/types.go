package sql

type Operation uint8

const (
	Select Operation = iota
	SelectOne
	Count
	Insert
	Update
	Delete
)

type Stmt interface {
	GetStmtWithArgs(operation Operation) (string, []any)
}

type StmtModel interface {
	GetStmtWithArgsForModel(operation Operation, model any) (string, []any)
}

type StmtSelect interface {
	Stmt
	GetModelPointers(operation Operation, model any) []any
}
