package cmd

import (
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
)

// TestBenchmarkContentIsViolationFree verifies that generateComplexMarkdown
// produces zero lint errors under benchmarkConfig, ensuring the benchmark
// measures rule scan cost rather than reporting-path allocations.
func TestBenchmarkContentIsViolationFree(t *testing.T) {
	content := generateComplexMarkdown(1000)
	cfg := benchmarkConfig()
	lint := linter.New(cfg)

	errs, _, _ := lint.LintContent("benchmark.md", content)
	if len(errs) != 0 {
		t.Errorf("expected 0 violations, got %d:", len(errs))
		for _, e := range errs {
			t.Errorf("  line %d [%s]: %s", e.Line, e.Rule, e.Message)
		}
	}
}
