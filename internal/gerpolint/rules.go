package gerpolint

import (
	"fmt"
	"strings"
)

// RuleID is the stable identifier used in diagnostic categories and in
// inline directives (//gerpolint:disable=GPL001).
type RuleID string

const (
	RuleScalarTypeMismatch      RuleID = "GPL001"
	RuleVariadicElementMismatch RuleID = "GPL002"
	RuleStringOnlyOperator      RuleID = "GPL003"
	RuleUnresolvedFieldPointer  RuleID = "GPL004"
	RuleAnyTypedArgument        RuleID = "GPL005"
	RuleDirectiveUnknown        RuleID = "GPL-DIRECTIVE-UNKNOWN"
)

var allRules = []RuleID{
	RuleScalarTypeMismatch,
	RuleVariadicElementMismatch,
	RuleStringOnlyOperator,
	RuleUnresolvedFieldPointer,
	RuleAnyTypedArgument,
}

func knownRule(r RuleID) bool {
	for _, k := range allRules {
		if k == r {
			return true
		}
	}
	return false
}

// Severity controls the behavior of rules whose applicability is
// context-dependent (unresolved field pointer, any-typed argument).
type Severity int

const (
	SeveritySkip Severity = iota
	SeverityWarn
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeveritySkip:
		return "skip"
	case SeverityWarn:
		return "warn"
	case SeverityError:
		return "error"
	}
	return "unknown"
}

func parseSeverity(s string) (Severity, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "skip":
		return SeveritySkip, nil
	case "warn":
		return SeverityWarn, nil
	case "error":
		return SeverityError, nil
	default:
		return 0, fmt.Errorf("severity: want skip|warn|error, got %q", s)
	}
}

// severityFlag implements flag.Value so Severity can be set from CLI.
type severityFlag struct{ target *Severity }

func (f severityFlag) String() string {
	if f.target == nil {
		return "skip"
	}
	return f.target.String()
}

func (f severityFlag) Set(s string) error {
	parsed, err := parseSeverity(s)
	if err != nil {
		return err
	}
	*f.target = parsed
	return nil
}
