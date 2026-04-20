package virtual

import (
	"context"
	"fmt"
	"reflect"
)

// FilterSpec describes the SQL fragment to use when filtering a virtual column for one operator.
// Construct it through one of the concrete variants — SQL, Bound, SQLArgs, Match, or Func.
type FilterSpec interface {
	isFilterSpec()
}

// SQL is a ready-made SQL fragment that does not depend on the user value or any extra
// bound parameters.
//
//	Filter(types.OperationEQ, virtual.SQL("EXISTS (SELECT 1 FROM x WHERE ...)"))
type SQL string

// Bound is a SQL fragment that contains exactly one `?` placeholder, into which the
// user-supplied value is bound.
//
//	Filter(types.OperationGT, virtual.Bound{SQL: "SUM(amount) > ?"})
type Bound struct {
	SQL string
}

// SQLArgs is a SQL fragment with explicit bound parameters; the user value is NOT bound.
// Useful when the predicate references constants or context that the column owner already has.
//
//	Filter(types.OperationEQ, virtual.SQLArgs{
//	    SQL:  "computed_at BETWEEN ? AND ?",
//	    Args: []any{from, to},
//	})
type SQLArgs struct {
	SQL  string
	Args []any
}

// Match discriminates on the user value: the first Case whose Value is equal (per
// reflect.DeepEqual) to the user value wins; if nothing matches, Default is used.
// A nil Default with no matching case yields an error from Apply.
//
//	Filter(types.OperationEQ, virtual.Match{
//	    Cases: []virtual.MatchCase{
//	        {Value: true,  Spec: virtual.SQL("EXISTS (...)")},
//	        {Value: false, Spec: virtual.SQL("NOT EXISTS (...)")},
//	    },
//	    Default: virtual.SQL("FALSE"),
//	})
type Match struct {
	Cases   []MatchCase
	Default FilterSpec
}

// MatchCase pairs a literal value with the spec to use when the user value matches it.
type MatchCase struct {
	Value any
	Spec  FilterSpec
}

// Func is the escape hatch for ctx-aware logic (multi-tenant, dynamic SQL). All
// other variants are static and can be tested by simple struct comparison.
//
//	Filter(types.OperationEQ, virtual.Func(func(ctx context.Context, v any) (string, []any, error) {
//	    tid := ctx.Value(tenantKey).(int)
//	    return "x.tenant_id = ? AND x.flag = ?", []any{tid, v}, nil
//	}))
type Func func(ctx context.Context, value any) (sql string, args []any, err error)

func (SQL) isFilterSpec()     {}
func (Bound) isFilterSpec()   {}
func (SQLArgs) isFilterSpec() {}
func (Match) isFilterSpec()   {}
func (Func) isFilterSpec()    {}

// compileFilter converts a FilterSpec into the args-based filter callback consumed by
// types.SQLFilterManager.AddFilterFnArgs.
func compileFilter(spec FilterSpec) func(context.Context, any) (string, []any, error) {
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
			cases[i] = compiled{key: c.Value, fn: compileFilter(c.Spec)}
		}
		var def func(context.Context, any) (string, []any, error)
		if s.Default != nil {
			def = compileFilter(s.Default)
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
			return "", nil, fmt.Errorf("virtual.Match: no case matched value %v and Default is nil", value)
		}
	case Func:
		return s
	}
	return func(context.Context, any) (string, []any, error) {
		return "", nil, fmt.Errorf("virtual: unknown FilterSpec type %T", spec)
	}
}
