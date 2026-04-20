package gerpolint

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// varState is the per-variable accumulator built by buildVarInits. `elts`
// grows as we apply recognized assignments in source order; `blocked`
// becomes sticky as soon as we see a use we cannot reason about.
type varState struct {
	elts    []ast.Expr
	blocked bool
	seeded  bool
}

// varInitIndex maps a local slice variable to the element expressions that
// flow into it. It recognizes two construction patterns:
//
//   - single-site composite literal:      xs := []any{a, b, c}
//   - append-chain starting from empty:   var xs []any
//     xs = append(xs, a)
//     xs = append(xs, b, c)
//
// Anything outside these patterns (function returns, `&xs`, `xs[i] = …`,
// reassignment from another slice) blocks the variable — `lookup` returns
// `false`, and the caller falls back to the GPL005 warning.
type varInitIndex struct {
	states map[*types.Var]*varState
}

func (vi *varInitIndex) lookup(v *types.Var) ([]ast.Expr, bool) {
	if vi == nil || v == nil {
		return nil, false
	}
	s, ok := vi.states[v]
	if !ok || s.blocked || !s.seeded {
		return nil, false
	}
	return s.elts, true
}

func (vi *varInitIndex) stateOf(v *types.Var) *varState {
	s, ok := vi.states[v]
	if !ok {
		s = &varState{}
		vi.states[v] = s
	}
	return s
}

func (vi *varInitIndex) block(v *types.Var) {
	if v == nil {
		return
	}
	s := vi.stateOf(v)
	s.blocked = true
	s.elts = nil
}

func (vi *varInitIndex) seed(v *types.Var, elts []ast.Expr) {
	s := vi.stateOf(v)
	if s.blocked {
		return
	}
	s.elts = append(s.elts[:0], elts...)
	s.seeded = true
}

func (vi *varInitIndex) appendElts(v *types.Var, elts []ast.Expr) {
	s := vi.stateOf(v)
	if s.blocked {
		return
	}
	s.elts = append(s.elts, elts...)
	s.seeded = true
}

// buildVarInits walks each file twice: pass 1 blocks any variable whose
// address is taken or whose elements are overwritten by index; pass 2
// applies composite-literal seeds and `append(v, …)` chains in source
// order. Any unrecognized assignment blocks the variable, so the
// accumulator never carries stale data past an opaque mutation.
func buildVarInits(pass *analysis.Pass) *varInitIndex {
	idx := &varInitIndex{states: map[*types.Var]*varState{}}

	varOf := func(expr ast.Expr) *types.Var {
		id, ok := expr.(*ast.Ident)
		if !ok {
			return nil
		}
		if v, ok := pass.TypesInfo.Defs[id].(*types.Var); ok {
			return v
		}
		if v, ok := pass.TypesInfo.Uses[id].(*types.Var); ok {
			return v
		}
		return nil
	}

	// Pass 1 — find disqualifying uses.
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.UnaryExpr:
				if node.Op == token.AND {
					idx.block(varOf(node.X))
				}
			case *ast.AssignStmt:
				for _, lhs := range node.Lhs {
					if ie, ok := lhs.(*ast.IndexExpr); ok {
						idx.block(varOf(ie.X))
					}
				}
			case *ast.IncDecStmt:
				if ie, ok := node.X.(*ast.IndexExpr); ok {
					idx.block(varOf(ie.X))
				}
			}
			return true
		})
	}

	apply := func(lhs, rhs ast.Expr) {
		v := varOf(lhs)
		if v == nil {
			return
		}
		if s := idx.states[v]; s != nil && s.blocked {
			return
		}

		if elts, ok := compositeSliceElts(rhs); ok {
			idx.seed(v, elts)
			return
		}

		if call, ok := rhs.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "append" &&
				len(call.Args) >= 1 && varOf(call.Args[0]) == v {
				if call.Ellipsis != token.NoPos {
					// Only inline-literal spreads are resolvable today:
					// `append(v, []any{a, b}...)`. Anything else (a named
					// slice variable, a function call) blocks.
					if len(call.Args) == 2 {
						if elts, ok := compositeSliceElts(call.Args[1]); ok {
							idx.appendElts(v, elts)
							return
						}
					}
					idx.block(v)
					return
				}
				idx.appendElts(v, call.Args[1:])
				return
			}
		}

		idx.block(v)
	}

	// Pass 2 — apply assignments in source order.
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt:
				if len(node.Lhs) != len(node.Rhs) {
					return true
				}
				for i, lhs := range node.Lhs {
					if _, ok := lhs.(*ast.IndexExpr); ok {
						continue
					}
					apply(lhs, node.Rhs[i])
				}
			case *ast.ValueSpec:
				for i, name := range node.Names {
					if i < len(node.Values) {
						apply(name, node.Values[i])
					}
				}
			}
			return true
		})
	}

	return idx
}

// compositeSliceElts returns the element expressions of a slice composite
// literal, unwrapping key-value forms (`[]any{0: x, 1: y}` → `[x, y]`). It
// returns false for non-literal expressions and for literals whose type is
// not a slice.
func compositeSliceElts(expr ast.Expr) ([]ast.Expr, bool) {
	cl, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}
	if !isSliceTypeExpr(cl.Type) {
		return nil, false
	}
	out := make([]ast.Expr, 0, len(cl.Elts))
	for _, e := range cl.Elts {
		if kv, ok := e.(*ast.KeyValueExpr); ok {
			out = append(out, kv.Value)
			continue
		}
		out = append(out, e)
	}
	return out, true
}

// isSliceTypeExpr tells a `[]T{...}` literal apart from map / array / struct
// literals; we only inline-expand slices.
func isSliceTypeExpr(e ast.Expr) bool {
	switch t := e.(type) {
	case *ast.ArrayType:
		return t.Len == nil
	case *ast.Ident:
		// Named type — resolution happens via types.Info at the call site.
		return false
	}
	return false
}

// spreadElements returns the element expressions backing `arg` when it is
// used as a variadic spread — either an inline composite literal or an
// identifier bound to a tracked accumulator via `buildVarInits`. Returns
// false when the source cannot be located statically.
func spreadElements(pass *analysis.Pass, vi *varInitIndex, arg ast.Expr) ([]ast.Expr, bool) {
	if elts, ok := compositeSliceElts(arg); ok {
		return elts, true
	}
	id, ok := arg.(*ast.Ident)
	if !ok {
		return nil, false
	}
	v, _ := pass.TypesInfo.Uses[id].(*types.Var)
	return vi.lookup(v)
}
