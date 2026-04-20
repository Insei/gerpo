// Package gerpolintplugin registers the gerpolint analyzer as a
// golangci-lint v2 module plugin. End users point golangci-lint at this
// package via .custom-gcl.yml; golangci-lint's `custom` subcommand builds
// a bespoke binary with the analyzer embedded.
package gerpolintplugin

import (
	"fmt"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/insei/gerpo/gerpolint"
)

func init() {
	register.Plugin("gerpolint", New)
}

// Settings is the YAML shape consumed by golangci-lint under
// `linters-settings.custom.gerpolint.settings`.
//
//	linters-settings:
//	  custom:
//	    gerpolint:
//	      type: module
//	      settings:
//	        unresolved-field: skip    # skip | warn | error
//	        any-arg: warn             # skip | warn | error
//	        disabled-rules: []        # [GPL001, GPL002, …]
type Settings struct {
	UnresolvedField string   `json:"unresolved-field" yaml:"unresolved-field"`
	AnyArg          string   `json:"any-arg"          yaml:"any-arg"`
	DisabledRules   []string `json:"disabled-rules"   yaml:"disabled-rules"`
}

// New is the register.NewPlugin factory. It decodes the YAML settings into
// a fresh Analyzer and applies them via the analyzer's Flags, reusing the
// same CLI semantics as the standalone binary.
func New(raw any) (register.LinterPlugin, error) {
	cfg, err := register.DecodeSettings[Settings](raw)
	if err != nil {
		return nil, fmt.Errorf("gerpolint: decode settings: %w", err)
	}
	a := gerpolint.NewAnalyzer()
	if cfg.UnresolvedField != "" {
		if err := a.Flags.Set("unresolved-field", cfg.UnresolvedField); err != nil {
			return nil, fmt.Errorf("gerpolint: unresolved-field: %w", err)
		}
	}
	if cfg.AnyArg != "" {
		if err := a.Flags.Set("any-arg", cfg.AnyArg); err != nil {
			return nil, fmt.Errorf("gerpolint: any-arg: %w", err)
		}
	}
	if len(cfg.DisabledRules) > 0 {
		if err := a.Flags.Set("disabled-rules", strings.Join(cfg.DisabledRules, ",")); err != nil {
			return nil, fmt.Errorf("gerpolint: disabled-rules: %w", err)
		}
	}
	return &plugin{analyzer: a}, nil
}

type plugin struct {
	analyzer *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.analyzer}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
