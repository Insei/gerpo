package databasesql

import "github.com/insei/gerpo/executor/adapters/placeholder"

type Option interface {
	apply(*dbWrap)
}

type optionFn func(*dbWrap)

func (o optionFn) apply(db *dbWrap) {
	o(db)
}

// WithPlaceholder sets a custom placeholder format for the database.
func WithPlaceholder(format placeholder.PlaceholderFormat) Option {
	return optionFn(func(db *dbWrap) {
		db.placeholder = format
	})
}
