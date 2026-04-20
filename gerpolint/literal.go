package gerpolint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// varInitIndex maps a local `[]any`-typed variable to the elements of the
// composite literal that initialized it, provided the variable has exactly
// one such assignment. It lets `checkVariadic` recover static element types
// from a pattern like `xs := []any{a, b, c}; ... In(xs...)`.
type varInitIndex struct {
	inits map[*types.Var][]ast.Expr
	// A variable seen assigned more than once, or assigned something other
	// than a usable composite literal, is parked here so we don't attempt
	// to reason about it. Presence in `blocked` overrides `inits`.
	blocked map[*types.Var]bool
}

func (vi *varInitIndex) lookup(v *types.Var) ([]ast.Expr, bool) {
	if vi == nil || v == nil {
		return nil, false
	}
	if vi.blocked[v] {
		return nil, false
	}
	elts, ok := vi.inits[v]
	return elts, ok
}

// buildVarInits walks every file in the pass once, recording single-site
// composite-literal assignments to local variables. Multiple assignments
// (or non-literal initializers after the first literal) block the variable.
func buildVarInits(pass *analysis.Pass) *varInitIndex {
	idx := &varInitIndex{
		inits:   map[*types.Var][]ast.Expr{},
		blocked: map[*types.Var]bool{},
	}

	record := func(v *types.Var, rhs ast.Expr) {
		if v == nil || idx.blocked[v] {
			return
		}
		elts, ok := compositeSliceElts(rhs)
		if !ok {
			if _, seen := idx.inits[v]; seen {
				delete(idx.inits, v)
				idx.blocked[v] = true
			}
			return
		}
		if _, seen := idx.inits[v]; seen {
			delete(idx.inits, v)
			idx.blocked[v] = true
			return
		}
		idx.inits[v] = elts
	}

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt:
				if len(node.Lhs) != len(node.Rhs) {
					// e.g. `a, b := fn()` — we can't attribute RHS elements per var.
					return true
				}
				for i, lhs := range node.Lhs {
					id, ok := lhs.(*ast.Ident)
					if !ok {
						continue
					}
					obj := pass.TypesInfo.Defs[id]
					if obj == nil {
						obj = pass.TypesInfo.Uses[id]
					}
					v, _ := obj.(*types.Var)
					record(v, node.Rhs[i])
				}
			case *ast.ValueSpec:
				for i, name := range node.Names {
					if i >= len(node.Values) {
						continue
					}
					v, _ := pass.TypesInfo.Defs[name].(*types.Var)
					record(v, node.Values[i])
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
// identifier bound to one via `buildVarInits`. Returns false when the
// source cannot be located statically.
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
