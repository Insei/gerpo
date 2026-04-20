package types

import (
	"context"

	"github.com/insei/fmap/v3"
)

type SQLAction string

const (
	SQLActionSelect = SQLAction("select")
	SQLActionInsert = SQLAction("insert")
	SQLActionGroup  = SQLAction("group")
	SQLActionUpdate = SQLAction("update")
	SQLActionSort   = SQLAction("sort")
)

// Column is an interface representing a table column within a database context.
// It supports methods related to SQL actions, column metadata, and SQL conversion.
type Column interface {
	SQLFilterGetter

	// IsAllowedAction determines if the specified SQLAction is permitted for the column and returns true if allowed.
	IsAllowedAction(a SQLAction) bool

	// GetAllowedActions returns a list of SQLActions that are permitted for this column.
	GetAllowedActions() []SQLAction

	// ToSQL generates the SQL representation of the column.
	ToSQL(ctx context.Context) string

	// GetPtr retrieves a pointer to the field of the provided model corresponding to this column.
	GetPtr(model any) any

	// GetField retrieves the associated fmap.Field for the column.
	GetField() fmap.Field

	// Name returns the name of the column as a string and a boolean indicating whether the name is valid or exists.
	Name() (string, bool)

	// Table returns the name of the table associated with the column and a boolean indicating success or failure of the retrieval.
	Table() (string, bool)

	// IsAggregate reports whether the column represents an aggregate expression (SUM, COUNT, ...).
	// WhereBuilder rejects WHERE conditions on aggregate columns unless the operator has an explicit
	// filter override (see HasFilterOverride).
	IsAggregate() bool

	// HasFilterOverride reports whether a custom filter was registered for the operation
	// (typically through virtual.Filter). Auto-derived filters return false.
	HasFilterOverride(op Operation) bool
}

// ColumnsGetter is an interface for retrieving a list of Column objects representing database table columns.
type ColumnsGetter interface {

	// GetColumns retrieves a list of Column objects representing database table columns.
	GetColumns() []Column
}

// ExecutionColumns represents an interface to manage and interact with a collection of database execution columns.
// It provides functionality to exclude columns, retrieve all columns, fetch columns by field pointers, and extract model data.
type ExecutionColumns interface {

	// Exclude removes the specified columns from the existing collection of execution columns, effectively excluding them from usage.
	Exclude(...Column)

	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(cols ...Column)

	// GetAll retrieves and returns all the columns contained within the execution columns as a slice.
	GetAll() []Column

	// GetByFieldPtr retrieves a Column based on the provided model and field pointer.
	// The method allows fetching specific columns related to the field in the execution context.
	GetByFieldPtr(model any, fieldPtr any) (Column, error)

	// GetModelPointers retrieves a slice of pointers to the fields of the given model based on the current execution columns.
	GetModelPointers(model any) []any

	// GetModelValues retrieves the values of the model's fields mapped to the execution columns and returns them as a slice.
	GetModelValues(model any) []any
}

// ColumnsStorage defines an interface for managing a collection of database columns.
type ColumnsStorage interface {

	// AsSlice returns all stored columns as a slice of type Column.
	AsSlice() []Column

	// NewExecutionColumns creates a new ExecutionColumns instance for the specified SQLAction within the provided context.
	NewExecutionColumns(ctx context.Context, action SQLAction) ExecutionColumns

	// GetByFieldPtr retrieves a Column by using the provided model and field pointer, returning an error if the Column is not found.
	GetByFieldPtr(model any, fieldPtr any) (Column, error)

	// Get checks if the specified field exists and returns the corresponding Column along with a boolean indicating success.
	Get(f fmap.Field) (Column, bool)

	// Add adds a new column to the storage, incorporating it into the collection of managed columns.
	Add(column Column)
}

type Operation string

const (
	// Shared operations

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

	// String filter operations.

	// OperationContains matches rows whose field contains the value as a substring (LIKE '%value%').
	OperationContains = Operation("contains")

	// OperationNotContains matches rows whose field does not contain the value as a substring.
	OperationNotContains = Operation("not_contains")

	// OperationStartsWith matches rows whose field begins with the value (LIKE 'value%').
	OperationStartsWith = Operation("starts_with")

	// OperationNotStartsWith matches rows whose field does not begin with the value.
	OperationNotStartsWith = Operation("not_starts_with")

	// OperationEndsWith matches rows whose field ends with the value (LIKE '%value').
	OperationEndsWith = Operation("ends_with")

	// OperationNotEndsWith matches rows whose field does not end with the value.
	OperationNotEndsWith = Operation("not_ends_with")

	// Case-insensitive variants of the string filter operations.
	// They are selected automatically by WhereOperation methods when ignoreCase=true is passed.

	// OperationContainsIgnoreCase is the case-insensitive form of OperationContains.
	OperationContainsIgnoreCase = Operation("contains_ic")

	// OperationNotContainsIgnoreCase is the case-insensitive form of OperationNotContains.
	OperationNotContainsIgnoreCase = Operation("not_contains_ic")

	// OperationStartsWithIgnoreCase is the case-insensitive form of OperationStartsWith.
	OperationStartsWithIgnoreCase = Operation("starts_with_ic")

	// OperationNotStartsWithIgnoreCase is the case-insensitive form of OperationNotStartsWith.
	OperationNotStartsWithIgnoreCase = Operation("not_starts_with_ic")

	// OperationEndsWithIgnoreCase is the case-insensitive form of OperationEndsWith.
	OperationEndsWithIgnoreCase = Operation("ends_with_ic")

	// OperationNotEndsWithIgnoreCase is the case-insensitive form of OperationNotEndsWith.
	OperationNotEndsWithIgnoreCase = Operation("not_ends_with_ic")
)

type OrderDirection string

const (
	OrderDirectionASC  = OrderDirection("ASC")
	OrderDirectionDESC = OrderDirection("DESC")
)

// SQLFilterManager manages operations and corresponding SQL generation functions for filtering.
// It allows adding custom filter functions for specific operations and retrieving available operations.
// This interface extends SQLFilterGetter for retrieving filter details and available operations.
type SQLFilterManager interface {

	// AddFilterFn registers a custom SQL generation function for a specific operation.
	// The boolean return value indicates whether the user value should be appended as a single
	// bound argument; for slice values, expansion is handled by the consumer (WhereBuilder).
	// Internally adapts to the args-based shape used by GetFilterFn.
	AddFilterFn(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, bool))

	// AddFilterFnArgs registers a filter that returns the bound arguments explicitly.
	// Used by callers that need to bind constants alongside (or instead of) the user value
	// — e.g. virtual columns with Compute(sql, args...) or Filter(op, virtual.SQLArgs{...}).
	AddFilterFnArgs(operation Operation, sqlGenFn func(ctx context.Context, value any) (string, []any, error))
	SQLFilterGetter
}

// SQLFilterGetter defines an interface for managing SQL filter operations and their corresponding logic.
type SQLFilterGetter interface {

	// GetFilterFn retrieves a filter function for the specified operation and indicates its availability.
	// The returned function yields the SQL fragment and the slice of bound arguments to append.
	GetFilterFn(operation Operation) (func(ctx context.Context, value any) (string, []any, error), bool)

	// GetAvailableFilterOperations retrieves a list of filter operations that are supported and available for use.
	GetAvailableFilterOperations() []Operation

	// IsAvailableFilterOperation checks whether the specified operation is in the list of available filter operations.
	IsAvailableFilterOperation(operation Operation) bool
}

type WhereOption func(op Operation) Operation

// WhereOperation defines an interface to apply various conditional operations for building queries.
type WhereOperation interface {

	// EQ applies an equality condition to the target field, comparing it with the specified value.
	EQ(val any) ANDOR

	// NEQ applies a "not equal to" condition to the query, comparing the field with the provided value.
	// It returns an ANDOR interface to chain further logical conditions.
	NEQ(val any) ANDOR

	// Contains applies a "contains" condition on the field with the provided value and returns an ANDOR for chaining.
	Contains(val any, ignoreCase ...bool) ANDOR

	// NotContains applies a "not contains" condition on the field with the provided value and returns an ANDOR for chaining.
	NotContains(val any, ignoreCase ...bool) ANDOR

	// StartsWith applies a "begins with" condition on the field with the provided value and returns an ANDOR for chaining.
	StartsWith(val any, ignoreCase ...bool) ANDOR

	// NotStartsWith applies a "not begins with" condition on the field with the provided value and returns an ANDOR for chaining.
	NotStartsWith(val any, ignoreCase ...bool) ANDOR

	// EndsWith applies an "ends with" condition on the field with the provided value and returns an ANDOR for chaining.
	EndsWith(val any, ignoreCase ...bool) ANDOR

	// NotEndsWith applies a "not ends with" condition on the field with the provided value and returns an ANDOR for chaining.
	NotEndsWith(val any, ignoreCase ...bool) ANDOR

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
