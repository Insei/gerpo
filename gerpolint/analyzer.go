// Package gerpolint is a static analyzer that catches type mismatches
// between gerpo field pointers and filter arguments:
//
//	h.Where().Field(&m.Age).EQ("18")  // reported: int vs string
//	h.Where().Field(&m.Age).Contains(...) // reported: Contains on non-string field
//
// The rule, by design: if the struct field is T, the argument must be T; if
// the struct field is *T, the argument may be T or *T.
package gerpolint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	gerpoTypesPkg   = "github.com/insei/gerpo/types"
	fieldMethodName = "Field"
)

// Config holds the runtime configuration consumed by run(); it is wired to
// Analyzer.Flags in NewAnalyzer().
type Config struct {
	Unresolved Severity
	AnyArg     Severity
	Disabled   ruleList
}

type ruleList []string

func (r *ruleList) String() string { return strings.Join(*r, ",") }
func (r *ruleList) Set(v string) error {
	if v == "" {
		*r = nil
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	*r = out
	return nil
}

// NewAnalyzer returns a new Analyzer instance with dedicated flag state.
// Tests use this to configure severities per-subtest; production callers use
// the package-level Analyzer variable.
func NewAnalyzer() *analysis.Analyzer {
	cfg := &Config{
		Unresolved: SeveritySkip,
		AnyArg:     SeverityWarn,
	}
	a := &analysis.Analyzer{
		Name:     "gerpolint",
		Doc:      "check type compatibility of arguments to gerpo WHERE filter operators",
		URL:      "https://github.com/insei/gerpo",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
	a.Flags.Var(severityFlag{&cfg.Unresolved}, "unresolved-field",
		"when Field(ptr) cannot be resolved to a concrete field: skip|warn|error (default skip)")
	a.Flags.Var(severityFlag{&cfg.AnyArg}, "any-arg",
		"when an argument's static type is `any`: skip|warn|error (default warn)")
	a.Flags.Var(&cfg.Disabled, "disabled-rules",
		"comma-separated list of rule IDs to disable entirely")
	return a
}

// Analyzer is the default singleton for singlechecker / golangci-lint.
var Analyzer = NewAnalyzer()

func run(pass *analysis.Pass, cfg *Config) (any, error) {
	disabled := make(map[RuleID]bool, len(cfg.Disabled))
	for _, id := range cfg.Disabled {
		disabled[RuleID(id)] = true
	}

	directives := make(map[*ast.File]*directiveIndex, len(pass.Files))
	for _, f := range pass.Files {
		directives[f] = buildDirectives(pass.Fset, f)
	}

	// Surface bad directive IDs once per occurrence.
	for _, f := range pass.Files {
		dx := directives[f]
		for _, e := range dx.unknownIDs {
			if disabled[RuleDirectiveUnknown] {
				continue
			}
			pass.Report(analysis.Diagnostic{
				Pos:      e.pos,
				Message:  fmt.Sprintf("%s: unknown rule id in directive: %s", RuleDirectiveUnknown, e.id),
				Category: string(RuleDirectiveUnknown),
			})
		}
	}

	// Build a fast lookup: token.File → *ast.File. Needed to locate the
	// correct directive index for a diagnostic's position.
	fileByTokenFile := make(map[*token.File]*ast.File, len(pass.Files))
	for _, f := range pass.Files {
		if tf := pass.Fset.File(f.Pos()); tf != nil {
			fileByTokenFile[tf] = f
		}
	}

	fileOf := func(pos token.Pos) *ast.File {
		if tf := pass.Fset.File(pos); tf != nil {
			return fileByTokenFile[tf]
		}
		return nil
	}

	report := func(rule RuleID, pos token.Pos, format string, args ...any) {
		if disabled[rule] {
			return
		}
		if f := fileOf(pos); f != nil {
			if dx := directives[f]; dx != nil && dx.disabled(rule, pos, pass.Fset) {
				return
			}
		}
		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Message:  fmt.Sprintf(format, args...),
			Category: string(rule),
		})
	}

	varInits := buildVarInits(pass)

	inspec := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspec.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		op, known := operators[fun.Sel.Name]
		if !known {
			return
		}
		if !isGerpoMethod(pass, fun) {
			return
		}

		fieldCall, fieldArg, ok := findFieldCall(pass, fun.X)
		if !ok {
			return
		}
		_ = fieldCall

		fieldType, resolved := resolveFieldType(pass, fieldArg)
		if !resolved {
			if cfg.Unresolved != SeveritySkip {
				report(RuleUnresolvedFieldPointer, fieldArg.Pos(),
					"%s: cannot statically resolve field type for Field(...).%s(...)",
					RuleUnresolvedFieldPointer, op.name)
			}
			return
		}

		if op.category == catStringOnly && !isStringLikeField(fieldType) {
			report(RuleStringOnlyOperator, call.Pos(),
				"%s: %s is only applicable to string/*string fields, got %s",
				RuleStringOnlyOperator, op.name, displayType(fieldType))
			return
		}

		switch op.kind {
		case opScalar:
			if len(call.Args) == 1 {
				checkScalarArg(pass, report, op, fieldType, call.Args[0], cfg, RuleScalarTypeMismatch)
			}
		case opVariadic:
			checkVariadic(pass, report, op, fieldType, call, cfg, varInits)
		}
	})

	return nil, nil
}

// isGerpoMethod reports whether the selector references a method declared in
// gerpo's types package, regardless of which concrete or interface type the
// value was seen as at the call site.
func isGerpoMethod(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return false
	}
	obj := selection.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == gerpoTypesPkg
}

// findFieldCall walks the receiver chain of an operator call and, if it ends
// in a `.Field(arg)` call from gerpo's types package, returns that CallExpr
// and its single argument.
func findFieldCall(pass *analysis.Pass, recv ast.Expr) (*ast.CallExpr, ast.Expr, bool) {
	call, ok := recv.(*ast.CallExpr)
	if !ok {
		return nil, nil, false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, nil, false
	}
	if sel.Sel.Name != fieldMethodName {
		return nil, nil, false
	}
	if !isGerpoMethod(pass, sel) {
		return nil, nil, false
	}
	if len(call.Args) != 1 {
		return nil, nil, false
	}
	return call, call.Args[0], true
}

// resolveFieldType extracts the struct-field type T from a literal `&m.X`
// expression. Returns (nil, false) for anything more complex — variables,
// function calls, method expressions — which the caller handles according
// to the -unresolved-field flag.
func resolveFieldType(pass *analysis.Pass, expr ast.Expr) (types.Type, bool) {
	u, ok := expr.(*ast.UnaryExpr)
	if !ok || u.Op != token.AND {
		return nil, false
	}
	t := pass.TypesInfo.TypeOf(u)
	if t == nil {
		return nil, false
	}
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return nil, false
	}
	return ptr.Elem(), true
}

// reportFn is the in-closure diagnostic reporter used by the checkers.
type reportFn func(RuleID, token.Pos, string, ...any)

func checkScalarArg(pass *analysis.Pass, report reportFn, op operatorSpec, fieldType types.Type, arg ast.Expr, cfg *Config, rule RuleID) {
	tv, ok := pass.TypesInfo.Types[arg]
	if !ok || tv.Type == nil {
		return
	}
	argType := tv.Type
	if isEmptyInterface(argType) {
		if cfg.AnyArg != SeveritySkip {
			report(RuleAnyTypedArgument, arg.Pos(),
				"%s: argument to %s has static type `any`; static check skipped",
				RuleAnyTypedArgument, op.name)
		}
		return
	}
	if !isCompatible(fieldType, argType, tv.Value, isUntypedConst(pass, arg)) {
		report(rule, arg.Pos(),
			"%s: %s: argument type %s is not compatible with field type %s",
			rule, op.name, displayType(argType), displayType(fieldType))
	}
}

func checkVariadic(pass *analysis.Pass, report reportFn, op operatorSpec, fieldType types.Type, call *ast.CallExpr, cfg *Config, vi *varInitIndex) {
	// `In(xs...)` — single slice argument spread.
	if call.Ellipsis != token.NoPos && len(call.Args) == 1 {
		argType := pass.TypesInfo.TypeOf(call.Args[0])
		if argType == nil {
			return
		}
		slice, ok := argType.Underlying().(*types.Slice)
		if !ok {
			return
		}
		elem := slice.Elem()
		if isEmptyInterface(elem) {
			// Recover static types from the backing composite literal when
			// available — inline `[]any{a, b}...` or a single-assignment
			// `xs := []any{a, b}; In(xs...)`.
			if elts, ok := spreadElements(pass, vi, call.Args[0]); ok {
				for _, e := range elts {
					checkScalarArg(pass, report, op, fieldType, e, cfg, RuleVariadicElementMismatch)
				}
				return
			}
			if cfg.AnyArg != SeveritySkip {
				report(RuleAnyTypedArgument, call.Args[0].Pos(),
					"%s: %s(xs...) element type is `any`; static check skipped",
					RuleAnyTypedArgument, op.name)
			}
			return
		}
		if !isCompatible(fieldType, elem, nil, false) {
			report(RuleVariadicElementMismatch, call.Args[0].Pos(),
				"%s: %s: spread element type %s is not compatible with field type %s",
				RuleVariadicElementMismatch, op.name, displayType(elem), displayType(fieldType))
		}
		return
	}

	for _, a := range call.Args {
		checkScalarArg(pass, report, op, fieldType, a, cfg, RuleVariadicElementMismatch)
	}
}
