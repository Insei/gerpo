package sqlstmt

import (
	"strings"

	"github.com/insei/gerpo/types"
)

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

// collectReturning walks the full ColumnsStorage and returns columns that the
// user marked as IsReturned(action) — those that should appear in a RETURNING
// clause for the given write action (INSERT or UPDATE). Returning-eligible
// columns may be omitted from the action's execution-columns set (e.g. a UUID
// PK that is OmitOnInsert + ReturnedOnInsert), so we go through the storage
// rather than the action's filtered ExecutionColumns.
func collectReturning(storage types.ColumnsStorage, action types.SQLAction) []types.Column {
	var out []types.Column
	for _, c := range storage.AsSlice() {
		if c.IsReturned(action) {
			out = append(out, c)
		}
	}
	return out
}

// appendReturning writes ` RETURNING name1, name2` to sb when cols is non-empty.
// Columns whose Name() is unset (e.g. virtual) are skipped — RETURNING needs
// real column names, not expressions.
func appendReturning(sb *strings.Builder, cols []types.Column) {
	if len(cols) == 0 {
		return
	}
	first := true
	for _, c := range cols {
		name, ok := c.Name()
		if !ok || name == "" {
			continue
		}
		if first {
			sb.WriteString(" RETURNING ")
			first = false
		} else {
			sb.WriteString(", ")
		}
		sb.WriteString(name)
	}
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
