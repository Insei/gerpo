package column

import (
	"strings"

	"github.com/insei/gerpo/types"
)

type options struct {
	table           string
	alias           string
	name            string
	notAvailActions []types.SQLAction
	returnedActions []types.SQLAction
}

type Option interface {
	apply(c *options)
}

// optionFunc is a type that implements the Option interface.
type columnOptionFn func(c *options)

// apply implements the Option interface for columnOptionFn.
// It calls the underlying function with the given *options.
func (f columnOptionFn) apply(c *options) {
	f(c)
}

func WithAlias(alias string) Option {
	return columnOptionFn(func(c *options) {
		if len(alias) > 0 {
			c.alias = alias
		}
	})
}

func WithTable(table string) Option {
	return columnOptionFn(func(c *options) {
		if len(table) > 0 {
			c.table = table
		}
	})
}

func WithColumnName(name string) Option {
	return columnOptionFn(func(c *options) {
		if len(name) > 0 {
			c.name = strings.TrimSpace(name)
		}
	})
}

// WithOmitOnInsert excludes the column from every INSERT statement. Use cases:
// values the application never supplies on create (UpdatedAt set by a trigger
// or hook on UPDATE, DeletedAt set by soft-delete).
func WithOmitOnInsert() Option {
	return columnOptionFn(func(c *options) {
		c.notAvailActions = append(c.notAvailActions, types.SQLActionInsert)
	})
}

// WithOmitOnUpdate excludes the column from every UPDATE SET clause. Use cases:
// immutable values set once on create (CreatedAt, the PK after insertion).
func WithOmitOnUpdate() Option {
	return columnOptionFn(func(c *options) {
		c.notAvailActions = append(c.notAvailActions, types.SQLActionUpdate)
	})
}

// WithReturnedOnInsert marks the column to appear in the INSERT ... RETURNING
// clause. After the SQL runs, the value scanned from RETURNING lands back into
// the model field — useful for server-generated PKs (DEFAULT gen_random_uuid()),
// DB DEFAULTs (NOW()), or trigger-managed columns. Pair with WithOmitOnInsert
// (or ReadOnly) when the column is fully DB-side.
func WithReturnedOnInsert() Option {
	return columnOptionFn(func(c *options) {
		c.returnedActions = append(c.returnedActions, types.SQLActionInsert)
	})
}

// WithReturnedOnUpdate marks the column to appear in the UPDATE ... RETURNING
// clause. After the SQL runs, the scanned value lands back into the model
// field — useful for trigger-updated columns (updated_at, version counters).
func WithReturnedOnUpdate() Option {
	return columnOptionFn(func(c *options) {
		c.returnedActions = append(c.returnedActions, types.SQLActionUpdate)
	})
}
