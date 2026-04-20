package gerpolint

import (
	"go/ast"
	"go/token"
	"strings"
)

// directiveSet represents the rules disabled by a single directive.
// If `all` is true, every rule on the target scope is disabled.
type directiveSet struct {
	all   bool
	rules map[RuleID]struct{}
}

func (d *directiveSet) contains(rule RuleID) bool {
	if d == nil {
		return false
	}
	if d.all {
		return true
	}
	_, ok := d.rules[rule]
	return ok
}

// blockRange is a [start, end] line range with a disable applied. end == 0
// means "open block" — until end of file, if no matching //gerpolint:enable
// appeared.
type blockRange struct {
	start, end int
	rules      *directiveSet
}

type directiveError struct {
	pos token.Pos
	id  string
}

type directiveIndex struct {
	disabledByLine map[int]*directiveSet
	blockDisabled  []blockRange
	unknownIDs     []directiveError
}

func (dx *directiveIndex) disabled(rule RuleID, pos token.Pos, fset *token.FileSet) bool {
	if dx == nil {
		return false
	}
	line := fset.Position(pos).Line
	if s, ok := dx.disabledByLine[line]; ok && s.contains(rule) {
		return true
	}
	for _, br := range dx.blockDisabled {
		if line >= br.start && (br.end == 0 || line <= br.end) {
			if br.rules.contains(rule) {
				return true
			}
		}
	}
	return false
}

const directivePrefix = "//gerpolint:"

// buildDirectives scans the file's comments once, producing an index that
// can be queried per-diagnostic. Supported verbs:
//
//	//gerpolint:disable[=r1,r2]           — until //gerpolint:enable or EOF
//	//gerpolint:enable                    — close the most recent disable
//	//gerpolint:disable-line[=r1,r2]      — current line (trailing comment)
//	//gerpolint:disable-next-line[=r1,r2] — the line immediately after
func buildDirectives(fset *token.FileSet, f *ast.File) *directiveIndex {
	dx := &directiveIndex{disabledByLine: map[int]*directiveSet{}}

	var (
		openStart int
		openRules *directiveSet
		isOpen    bool
	)

	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if !strings.HasPrefix(c.Text, directivePrefix) {
				continue
			}
			verb, rest := splitVerb(c.Text[len(directivePrefix):])
			rules, unknown := parseRules(rest)
			if unknown != "" {
				dx.unknownIDs = append(dx.unknownIDs, directiveError{pos: c.Slash, id: unknown})
				continue
			}
			line := fset.Position(c.Slash).Line
			switch verb {
			case "disable":
				if !isOpen {
					isOpen = true
					openStart = line
					openRules = rules
				}
			case "enable":
				if isOpen {
					dx.blockDisabled = append(dx.blockDisabled, blockRange{
						start: openStart, end: line, rules: openRules,
					})
					isOpen = false
				}
			case "disable-line":
				dx.disabledByLine[line] = rules
			case "disable-next-line":
				dx.disabledByLine[line+1] = rules
			}
		}
	}

	if isOpen {
		dx.blockDisabled = append(dx.blockDisabled, blockRange{
			start: openStart, end: 0, rules: openRules,
		})
	}
	return dx
}

func splitVerb(s string) (verb, rest string) {
	for i, r := range s {
		if r == ' ' || r == '=' || r == '\t' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// parseRules interprets the `=r1,r2` (or empty / whitespace) tail after the
// verb. Returns the disable set or, on unknown ID, the offending token as
// the second return value.
func parseRules(s string) (*directiveSet, string) {
	s = strings.TrimLeft(s, " =\t")
	if s == "" {
		return &directiveSet{all: true}, ""
	}
	set := &directiveSet{rules: map[RuleID]struct{}{}}
	for _, raw := range strings.Split(s, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if !knownRule(RuleID(id)) {
			return nil, id
		}
		set.rules[RuleID(id)] = struct{}{}
	}
	return set, ""
}
