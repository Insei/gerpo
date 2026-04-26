package virtual

import (
	"context"

	"github.com/insei/gerpo/filters"
)

// FilterSpec, SQL, Bound, SQLArgs, Match, MatchCase, Func are aliases for the
// types in the filters package. The filter vocabulary now lives there so plain
// columns and the global Registry can use it; virtual keeps its old import
// path zero-cost.
type (
	FilterSpec = filters.FilterSpec
	SQL        = filters.SQL
	Bound      = filters.Bound
	SQLArgs    = filters.SQLArgs
	Match      = filters.Match
	MatchCase  = filters.MatchCase
	Func       = filters.Func
)

// compileFilter is retained as a thin wrapper because virtual.options uses it.
func compileFilter(spec FilterSpec) func(context.Context, any) (string, []any, error) {
	return filters.CompileFilter(spec)
}
