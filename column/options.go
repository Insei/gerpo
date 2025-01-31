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

func WithInsertProtection() Option {
	return columnOptionFn(func(c *options) {
		c.notAvailActions = append(c.notAvailActions, types.SQLActionInsert)
	})
}

func WithUpdateProtection() Option {
	return columnOptionFn(func(c *options) {
		c.notAvailActions = append(c.notAvailActions, types.SQLActionUpdate)
	})
}
