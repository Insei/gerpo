package databasesql

import "github.com/insei/gerpo/executor/adapters/placeholder"

// Option tunes how NewAdapter wires the underlying *sql.DB.
type Option interface {
	apply(*adapterConfig)
}

type optionFn func(*adapterConfig)

func (o optionFn) apply(cfg *adapterConfig) { o(cfg) }

// WithPlaceholder sets a custom placeholder format. The default is
// placeholder.Question (`?`); use placeholder.Dollar for PostgreSQL.
func WithPlaceholder(format placeholder.PlaceholderFormat) Option {
	return optionFn(func(cfg *adapterConfig) {
		cfg.placeholder = format
	})
}
