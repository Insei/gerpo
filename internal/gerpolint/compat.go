package gerpolint

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// isCompatible implements the gerpo filter-argument rule:
//   - field T    → argument must be assignable to T;
//   - field *T   → argument may be assignable to T or to *T.
//
// For untyped constants (literals, untyped-const identifiers), assignability
// is relaxed to value-level representability against the basic underlying of
// the target: `18` is representable as `type Age int`, but `3.14` is not
// representable as `int`, and `"a"` is not representable as `int`.
func isCompatible(fieldType, argType types.Type, constVal constant.Value, untyped bool) bool {
	targets := []types.Type{fieldType}
	if ptr, ok := fieldType.(*types.Pointer); ok {
		targets = append(targets, ptr.Elem())
	}

	for _, target := range targets {
		if b, ok := argType.(*types.Basic); ok && b.Kind() == types.UntypedNil {
			if types.AssignableTo(argType, target) {
				return true
			}
			continue
		}
		if untyped && constVal != nil && constVal.Kind() != constant.Unknown {
			if basic, ok := target.Underlying().(*types.Basic); ok {
				if constRepresentable(constVal, basic) {
					return true
				}
				continue
			}
			// Non-basic target (pointer, interface, struct): fall through
			// to standard assignability on the pointer-elem pass.
		}
		if types.AssignableTo(argType, target) {
			return true
		}
	}
	return false
}

// isStringLikeField returns true when the underlying type of the field (after
// dereferencing one level of pointer) is a string. Named types whose
// underlying is string — e.g. `type Alias string` — count as string-like.
func isStringLikeField(fieldType types.Type) bool {
	t := fieldType
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	return basic.Kind() == types.String
}

// isEmptyInterface reports whether t is `any` (equivalently `interface{}`).
// Used to classify arguments whose static type carries no compile-time info.
func isEmptyInterface(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return iface.NumMethods() == 0
}

// isUntypedConst reports whether the argument expression was originally an
// untyped constant, even after go/types defaulted it because the parameter
// is `any`. We inspect the AST because tv.Type alone cannot distinguish
// `EQ(K)` with `const K = 18` (untyped) from `EQ(K)` with `const K int = 18`
// (typed) after defaulting.
func isUntypedConst(pass *analysis.Pass, arg ast.Expr) bool {
	switch a := arg.(type) {
	case *ast.BasicLit:
		return true
	case *ast.ParenExpr:
		return isUntypedConst(pass, a.X)
	case *ast.UnaryExpr:
		return isUntypedConst(pass, a.X)
	case *ast.BinaryExpr:
		return isUntypedConst(pass, a.X) && isUntypedConst(pass, a.Y)
	case *ast.Ident:
		if a.Name == "nil" || a.Name == "true" || a.Name == "false" || a.Name == "iota" {
			return true
		}
		obj := pass.TypesInfo.ObjectOf(a)
		c, ok := obj.(*types.Const)
		if !ok {
			return false
		}
		b, ok := c.Type().(*types.Basic)
		if !ok {
			return false
		}
		return b.Info()&types.IsUntyped != 0
	}
	return false
}

// displayType renders a type for diagnostic messages.
func displayType(t types.Type) string {
	if t == nil {
		return "<unknown>"
	}
	return t.String()
}

// constRepresentable reports whether the constant value v can be represented
// as a value of the given basic type. Mirrors Go's spec-level representability
// rule (§Representability) using go/constant's Kind-conversion helpers.
func constRepresentable(v constant.Value, target *types.Basic) bool {
	info := target.Info()
	switch {
	case info&types.IsBoolean != 0:
		return v.Kind() == constant.Bool
	case info&types.IsString != 0:
		return v.Kind() == constant.String
	case info&types.IsInteger != 0:
		return constant.ToInt(v).Kind() == constant.Int
	case info&types.IsFloat != 0:
		return constant.ToFloat(v).Kind() == constant.Float
	case info&types.IsComplex != 0:
		return constant.ToComplex(v).Kind() == constant.Complex
	}
	return false
}
