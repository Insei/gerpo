package sqlstmt

import "github.com/insei/gerpo/types"

// mergeArgs concatenates positional argument slices in the order they appear
// in the generated SQL. JOIN arguments are emitted before the WHERE clause,
// so any bound JOIN values must precede WHERE values in the final []any.
//
// The function preserves nil semantics: if every input is empty the return
// value is nil so that callers can compare against the previous behavior
// without spurious empty-slice results.
func mergeArgs(slices ...[]any) []any {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	if total == 0 {
		return nil
	}
	out := make([]any, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

// columnArgsProvider is implemented by columns that contribute bound parameters
// every time their SQL expression is emitted (currently virtual columns built
// with Compute(sql, args...)).
type columnArgsProvider interface {
	SQLArgs() []any
}

// collectSelectArgs walks the columns in their SELECT order and accumulates
// any bound parameters those columns contribute through their SQL expression.
// Returns nil when no column contributes args.
func collectSelectArgs(cols []types.Column) []any {
	var out []any
	for _, c := range cols {
		ap, ok := c.(columnArgsProvider)
		if !ok {
			continue
		}
		args := ap.SQLArgs()
		if len(args) == 0 {
			continue
		}
		out = append(out, args...)
	}
	return out
}
