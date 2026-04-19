package sqlstmt

// mergeArgs concatenates positional argument slices in the order they appear
// in the generated SQL. JOIN arguments are emitted before the WHERE clause,
// so any bound JOIN values must precede WHERE values in the final []any.
//
// The function preserves nil semantics: if both inputs are empty the return
// value is nil so that callers can compare against the previous behavior
// without spurious empty-slice results.
func mergeArgs(left, right []any) []any {
	if len(left) == 0 {
		return right
	}
	if len(right) == 0 {
		return left
	}
	out := make([]any, 0, len(left)+len(right))
	out = append(out, left...)
	out = append(out, right...)
	return out
}
