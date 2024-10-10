package types

import (
	"context"
	"slices"
)

type Operation string

const (
	// OperationEQ is a constant of type Operation that represents the operation where the field is equal to the value
	OperationEQ = Operation("eq")

	// OperationNEQ is a constant of type Operation that represents the operation where the field is not equal to the value.
	OperationNEQ = Operation("neq")

	// OperationGT is a constant of type Operation that represents the operation where the field is greater than the value.
	OperationGT = Operation("gt")

	// OperationGTE is a constant of type Operation and represents the operation where the field is greater than or equal to the value.
	OperationGTE = Operation("gte")

	// OperationLT is a constant of type Operation and represents the operation where the field is less than the value.
	OperationLT = Operation("lt")

	// OperationLTE is a constant of type Operation that represents the operation where the field is less than or equal to the value.
	OperationLTE = Operation("lte")

	// OperationIN is a constant of type Operation that represents the operation where the field is in the specified values.
	OperationIN = Operation("in")

	// OperationNIN is a constant of type Operation that represents the operation where the field is not in the specified values.
	OperationNIN = Operation("nin")

	// OperationCT is a constant of type Operation that represents the operation where the field contains the value string.
	OperationCT = Operation("ct")
	// OperationNCT is a constant of type Operation that represents the operation where the field not contains the value string.
	OperationNCT = Operation("nct")
	// OperationEW is a constant of type Operation that represents the operation where the field ends with the value string.
	OperationEW = Operation("ew")

	// OperationNEW is a constant of type Operation. It represents the operation where the field not ends with the value string.
	OperationNEW = Operation("new")

	// OperationBW is a constant of type Operation. It represents the operation where the field begins with the value string.
	OperationBW = Operation("bw")

	// OperationNBW is a constant of type Operation that represents the operation where the field begins with the value string.
	OperationNBW = Operation("nbw")
)

type OrderDirection string

const (
	OrderDirectionASC  = OrderDirection("ASC")
	OrderDirectionDESC = OrderDirection("DESC")
)

var supportedOperations = []Operation{
	OperationEQ,
	OperationNEQ,
	OperationGT,
	OperationGTE,
	OperationLT,
	OperationLTE,
	OperationIN,
	OperationNIN,
	OperationCT,
	OperationNCT,
	OperationEW,
	OperationNEW,
	OperationBW,
	OperationNBW,
}

func IsSupportedOperation(op Operation) bool {
	return slices.Contains(supportedOperations, op)
}

type SQLFilterManager interface {
	AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool))
	SQLFilterGetter
}

type SQLFilterGetter interface {
	GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, bool, error), bool)
	GetAvailableFilterOperations() []Operation
	IsAvailableFilterOperation(operation Operation) bool
}

type WhereOperation[TModel any] interface {
	EQ(val any) ANDOR[TModel]
	NEQ(val any) ANDOR[TModel]
	CT(val any) ANDOR[TModel]
	NCT(val any) ANDOR[TModel]
	BW(val any) ANDOR[TModel]
	NBW(val any) ANDOR[TModel]
	EW(val any) ANDOR[TModel]
	NEW(val any) ANDOR[TModel]
	GT(val any) ANDOR[TModel]
	GTE(val any) ANDOR[TModel]
	LT(val any) ANDOR[TModel]
	LTE(val any) ANDOR[TModel]
}
type OrderOperation[TModel any] interface {
	DESC() OrderTarget[TModel]
	ASC() OrderTarget[TModel]
}

type OrderTarget[TModel any] interface {
	Field(fieldPtr any) OrderOperation[TModel]
}

type WhereTarget[TModel any] interface {
	Field(fieldPtr any) WhereOperation[TModel]
	Group(func(t WhereTarget[TModel])) ANDOR[TModel]
}

type ANDOR[TModel any] interface {
	OR() WhereTarget[TModel]
	AND() WhereTarget[TModel]
}

type InsertQuery interface {
}
