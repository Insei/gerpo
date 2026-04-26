// Package filters owns the project-wide WHERE-clause filter registry and the
// FilterSpec vocabulary used to override or extend stock SQL fragments. The
// registry exposes named buckets for built-in Go types (Bool, String, Numeric,
// Time, UUID) so callers can see which types ship with filters out of the box;
// custom types are added via Registry.Register.
package filters

import (
	"context"
	"fmt"
	"reflect"
)

// FilterSpec describes the SQL fragment to use when filtering by one operator.
// Construct it through one of the concrete variants — SQL, Bound, SQLArgs,
// Match, or Func.
type FilterSpec interface {
	isFilterSpec()
}

// SQL is a ready-made SQL fragment that does not depend on the user value or
// any extra bound parameters.
//
//	bucket.Override(types.OperationEQ, filters.SQL("EXISTS (SELECT 1 FROM x WHERE ...)"))
type SQL string

// Bound is a SQL fragment that contains exactly one `?` placeholder, into
// which the user-supplied value is bound.
//
//	bucket.Override(types.OperationGT, filters.Bound{SQL: "SUM(amount) > ?"})
type Bound struct {
	SQL string
}

// SQLArgs is a SQL fragment with explicit bound parameters; the user value is
// NOT bound. Useful when the predicate references constants or context that
// the column owner already has.
//
//	bucket.Override(types.OperationEQ, filters.SQLArgs{
//	    SQL:  "computed_at BETWEEN ? AND ?",
//	    Args: []any{from, to},
//	})
type SQLArgs struct {
	SQL  string
	Args []any
}

// Match discriminates on the user value: the first Case whose Value is equal
// (per reflect.DeepEqual) to the user value wins; if nothing matches, Default
// is used. A nil Default with no matching case yields an error.
//
//	bucket.Override(types.OperationEQ, filters.Match{
//	    Cases: []filters.MatchCase{
//	        {Value: true,  Spec: filters.SQL("EXISTS (...)")},
//	        {Value: false, Spec: filters.SQL("NOT EXISTS (...)")},
//	    },
//	    Default: filters.SQL("FALSE"),
//	})
type Match struct {
	Cases   []MatchCase
	Default FilterSpec
}

// MatchCase pairs a literal value with the spec to use when the user value
// matches it.
type MatchCase struct {
	Value any
	Spec  FilterSpec
}

// Func is the escape hatch for ctx-aware logic (multi-tenant, dynamic SQL).
// All other variants are static and can be tested by simple struct comparison.
//
//	bucket.Override(types.OperationEQ, filters.Func(func(ctx context.Context, v any) (string, []any, error) {
//	    tid := ctx.Value(tenantKey).(int)
//	    return "x.tenant_id = ? AND x.flag = ?", []any{tid, v}, nil
//	}))
type Func func(ctx context.Context, value any) (sql string, args []any, err error)

func (SQL) isFilterSpec()     {}
func (Bound) isFilterSpec()   {}
func (SQLArgs) isFilterSpec() {}
func (Match) isFilterSpec()   {}
func (Func) isFilterSpec()    {}

// CompileFilter converts a FilterSpec into the args-based filter callback
// consumed by types.SQLFilterManager. Exported because virtual.WithFilter
// re-uses the same machinery.
func CompileFilter(spec FilterSpec) func(context.Context, any) (string, []any, error) {
	switch s := spec.(type) {
	case SQL:
		sql := string(s)
		return func(_ context.Context, _ any) (string, []any, error) {
			return sql, nil, nil
		}
	case Bound:
		sql := s.SQL
		return func(_ context.Context, value any) (string, []any, error) {
			return sql, []any{value}, nil
		}
	case SQLArgs:
		sql := s.SQL
		args := append([]any(nil), s.Args...)
		return func(_ context.Context, _ any) (string, []any, error) {
			return sql, args, nil
		}
	case Match:
		type compiled struct {
			key any
			fn  func(context.Context, any) (string, []any, error)
		}
		cases := make([]compiled, len(s.Cases))
		for i, c := range s.Cases {
			cases[i] = compiled{key: c.Value, fn: CompileFilter(c.Spec)}
		}
		var def func(context.Context, any) (string, []any, error)
		if s.Default != nil {
			def = CompileFilter(s.Default)
		}
		return func(ctx context.Context, value any) (string, []any, error) {
			for _, c := range cases {
				if reflect.DeepEqual(c.key, value) {
					return c.fn(ctx, value)
				}
			}
			if def != nil {
				return def(ctx, value)
			}
			return "", nil, fmt.Errorf("filters.Match: no case matched value %v and Default is nil", value)
		}
	case Func:
		return s
	}
	return func(context.Context, any) (string, []any, error) {
		return "", nil, fmt.Errorf("filters: unknown FilterSpec type %T", spec)
	}
}
