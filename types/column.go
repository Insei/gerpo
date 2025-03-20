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
}

// ColumnsGetter is an interface for retrieving a list of Column objects representing database table columns.
type ColumnsGetter interface {

	// GetColumns retrieves a list of Column objects representing database table columns.
	GetColumns() []Column
}

// NewColumnBase creates a new instance of ColumnBase with the provided field, SQL generation function, and filters manager.
func NewColumnBase(field fmap.Field, toSQLFn func(ctx context.Context) string, filters SQLFilterManager) *ColumnBase {
	return &ColumnBase{
		Field:          field,
		ToSQL:          toSQLFn,
		AllowedActions: make([]SQLAction, 0),
		Filters:        filters,
		GetPtr: func(model any) any {
			return field.GetPtr(model)
		},
	}
}

// ColumnBase represents a base structure for defining and manipulating column-related behaviors in SQL operations.
type ColumnBase struct {
	Field          fmap.Field
	ToSQL          func(ctx context.Context) string
	AllowedActions []SQLAction
	Filters        SQLFilterManager
	GetPtr         func(model any) any
}

// IsAllowedAction determines if a given SQLAction is allowed for the column.
func (c *ColumnBase) IsAllowedAction(act SQLAction) bool {
	for _, a := range c.AllowedActions {
		if a == act {
			return true
		}
	}
	return false
}
