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

// SQLFilterManager manages operations and corresponding SQL generation functions for filtering.
// It allows adding custom filter functions for specific operations and retrieving available operations.
// This interface extends SQLFilterGetter for retrieving filter details and available operations.
type SQLFilterManager interface {

	// AddFilterFn registers a custom SQL generation function for a specific operation to handle filter generation logic.
	AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool))
	SQLFilterGetter
}

// SQLFilterGetter defines an interface for managing SQL filter operations and their corresponding logic.
type SQLFilterGetter interface {

	// GetFilterFn retrieves a filter function for the specified operation and indicates its availability.
	GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, bool, error), bool)

	// GetAvailableFilterOperations retrieves a list of filter operations that are supported and available for use.
	GetAvailableFilterOperations() []Operation

	// IsAvailableFilterOperation checks whether the specified operation is in the list of available filter operations.
	IsAvailableFilterOperation(operation Operation) bool
}

// WhereOperation defines an interface to apply various conditional operations for building queries.
type WhereOperation interface {

	// EQ applies an equality condition to the target field, comparing it with the specified value.
	EQ(val any) ANDOR

	// NEQ applies a "not equal to" condition to the query, comparing the field with the provided value.
	// It returns an ANDOR interface to chain further logical conditions.
	NEQ(val any) ANDOR

	// CT applies a "contains, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	CT(val any) ANDOR

	// NCT applies a "not contains, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	NCT(val any) ANDOR

	// BW applies a "begins with, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	BW(val any) ANDOR

	// NBW applies a "not begins with, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	NBW(val any) ANDOR

	// EW applies "ends with, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	EW(val any) ANDOR

	// NEW applies "not ends with, ignore case" condition on the field with the provided value and returns an ANDOR for chaining.
	NEW(val any) ANDOR

	// GT applies a greater-than (>) condition on the field with the provided value and returns an ANDOR for chaining.
	GT(val any) ANDOR

	// GTE applies a "greater than or equal to" (>=) condition on the field with the provided value and returns an ANDOR for chaining.
	GTE(val any) ANDOR

	// LT applies a "less than" (<) condition on the field with the provided value and returns an ANDOR for chaining.
	LT(val any) ANDOR

	// LTE applies a "less than or equal to" (<=) condition on the field with the provided value and returns an ANDOR for chaining.
	LTE(val any) ANDOR

	// IN applies the IN operation to filter records where the specified field matches any of the provided values.
	// Can accept slices in first argument.
	IN(vals ...any) ANDOR

	// NIN applies the NIN operation to filter records where the specified field not matches any of the provided values.
	// Can accept slices in first argument.
	NIN(vals ...any) ANDOR

	// OP applies a custom operation using the specified Operation type and value for conditional query construction.
	OP(operation Operation, val any) ANDOR
}

// OrderOperation represents an interface for defining order directives (ascending or descending) for query sorting.
type OrderOperation interface {

	// DESC specifies a descending order for a query and returns OrderTarget for further configuration.
	DESC() OrderTarget

	// ASC specifies the ascending order for a query and returns OrderTarget for further configuration.
	ASC() OrderTarget
}

// OrderTarget defines an interface for specifying an ordering operation in a query using fields or columns.
type OrderTarget interface {

	// Field specifies a field to be used for ordering operations and returns an OrderOperation for further configuration.
	Field(fieldPtr any) OrderOperation

	// Column specifies an order operation using the provided Column interface and returns an OrderOperation instance.
	Column(col Column) OrderOperation
}

// GroupTarget represents an interface for configuring grouping targets in a structured query or operation.
type GroupTarget interface {

	// Field adds one or more fields to the grouping target for configuring structured queries or operations.
	Field(fieldsPtr ...any) GroupTarget
}

// WhereTarget is an interface for building SQL WHERE clauses using column or field operations and logical groupings.
// It provides methods to specify conditions on individual columns or fields and allows grouping conditions with logic.
type WhereTarget interface {

	// Column specifies a condition for the given database column and returns a WhereOperation for applying SQL operations.
	Column(col Column) WhereOperation

	// Field allows specifying a field in the query for conditional operations. Input is a pointer to a struct field.
	Field(fieldPtr any) WhereOperation

	// Group starts a logical group for WHERE conditions and allows nesting conditions within the group.
	Group(func(t WhereTarget)) ANDOR
}

// ANDOR provides methods for chaining logical SQL conditions in WHERE clauses using AND and OR operators.
// The OR method allows adding an OR logical operator and continues building conditions.
// The AND method allows adding an AND logical operator and continues building conditions.
type ANDOR interface {

	// OR adds an OR logical operator to the WHERE clause and continues building conditions.
	OR() WhereTarget

	// AND adds a logical AND condition to the WHERE clause and continues the chain for building further conditions.
	AND() WhereTarget
}
