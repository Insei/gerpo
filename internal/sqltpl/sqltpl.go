// Package sqltpl holds the stock SQL fragment generators used by the WHERE
// clause builder and the filter registry. Generators take the column SQL
// reference (e.g. "users.id") and return a function that builds the operator
// fragment for one user value.
//
// The (string, bool) return shape mirrors the legacy sqlpart contract:
//   - sql:         the WHERE fragment ending with "?" (or "(?,?,?)" for In, …)
//     or a constant predicate ("1 = 1" / "1 = 2" / "IS NULL"),
//   - appendValue: whether the user value should be appended as a bound arg.
package sqltpl

import (
	"context"
	"strings"
	"unsafe"
)

// Generator is the shared shape of every fragment generator. A generator is
// produced once per (operator, column SQL) pair and then called per request
// with the user value.
type Generator func(query string) func(ctx context.Context, value any) (string, bool)

func EQ(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NULL", false
		}
		return query + " = ?", true
	}
}

func NotEQ(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NOT NULL", false
		}
		return query + " != ?", true
	}
}

func LT(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " < ?", true
	}
}

func LTE(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " <= ?", true
	}
}

func GT(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " > ?", true
	}
}

func GTE(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " >= ?", true
	}
}

// In/NotIn rely on the slice-of-any layout the where-builder feeds them; the
// unsafe access avoids reflect for the hot path. Empty slices collapse to a
// constant predicate so the rendered SQL stays valid without bound args.

func In(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if value == nil || len(*anyArr) == 0 {
			return "1 = 2", false
		}
		placeholders := strings.Repeat("?,", len(*anyArr))
		placeholders = placeholders[:len(placeholders)-1]
		return query + " IN (" + placeholders + ")", true
	}
}

func NotIn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if value == nil || len(*anyArr) == 0 {
			return "1 = 1", false
		}
		placeholders := strings.Repeat("?,", len(*anyArr))
		placeholders = placeholders[:len(placeholders)-1]
		return query + " NOT IN (" + placeholders + ")", true
	}
}

// LIKE-family: bound arg wrapped in CAST(? AS text) so PostgreSQL can infer
// the parameter type inside CONCAT. Works the same way on MySQL, so the
// fragment stays portable.

func Contains(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT('%', CAST(? AS text), '%')", true
	}
}

func NotContains(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT('%', CAST(? AS text), '%')", true
	}
}

func StartsWith(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT(CAST(? AS text), '%')", true
	}
}

func NotStartsWith(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT(CAST(? AS text), '%')", true
	}
}

func EndsWith(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT('%', CAST(? AS text))", true
	}
}

func NotEndsWith(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT('%', CAST(? AS text))", true
	}
}

// Case-insensitive "fold" variants — naming mirrors strings.EqualFold.

func EQFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NULL", false
		}
		return "LOWER(" + query + ") = LOWER(CAST(? AS text))", true
	}
}

func NotEQFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NOT NULL", false
		}
		return "LOWER(" + query + ") != LOWER(CAST(? AS text))", true
	}
}

func ContainsFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT('%', CAST(? AS text), '%'))", true
	}
}

func NotContainsFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT('%', CAST(? AS text), '%'))", true
	}
}

func StartsWithFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT(CAST(? AS text), '%'))", true
	}
}

func NotStartsWithFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT(CAST(? AS text), '%'))", true
	}
}

func EndsWithFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT('%', CAST(? AS text)))", true
	}
}

func NotEndsWithFold(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT('%', CAST(? AS text)))", true
	}
}
