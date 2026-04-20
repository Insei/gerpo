package gerpolint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCompat(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "compat")
}

func TestVariadic(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "variadic")
}

func TestStringOnly(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "stringonly")
}

func TestChains(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "chains")
}

func TestIdentity(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "identity")
}

func TestUnresolvedStrict(t *testing.T) {
	a := NewAnalyzer()
	if err := a.Flags.Set("unresolved-field", "error"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, analysistest.TestData(), a, "unresolved")
}

func TestUnresolvedDefault(t *testing.T) {
	// Default skip: no diagnostics expected. We assert by running on the
	// same fixture without -unresolved-field flag and expecting clean output
	// via a separate dir with no `// want` markers.
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "unresolved_skip")
}

func TestAnyArgWarn(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "anyarg")
}

func TestAnyArgSkip(t *testing.T) {
	a := NewAnalyzer()
	if err := a.Flags.Set("any-arg", "skip"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, analysistest.TestData(), a, "anyarg_skip")
}

func TestDirectives(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "directives")
}

func TestRegression(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "regression")
}
