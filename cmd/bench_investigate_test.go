package cmd

import (
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
)

// BenchmarkWith12Rules runs full linting WITHOUT no-trailing-spaces (simulates main).
func BenchmarkWith12Rules(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	cfg.Rules["no-trailing-spaces"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// BenchmarkWith13Rules runs full linting WITH no-trailing-spaces.
func BenchmarkWith13Rules(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// BenchmarkWith12Rules_XL runs full linting WITHOUT no-trailing-spaces (5000 sections).
func BenchmarkWith12Rules_XL(b *testing.B) {
	content := generateComplexMarkdown(5000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	cfg.Rules["no-trailing-spaces"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// BenchmarkWith13Rules_XL runs full linting WITH no-trailing-spaces (5000 sections).
func BenchmarkWith13Rules_XL(b *testing.B) {
	content := generateComplexMarkdown(5000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// --- Per-rule isolation benchmarks ---

func benchSingleRule(b *testing.B, ruleName string, sections int) {
	content := generateComplexMarkdown(sections)
	cfg := config.Default()
	for name := range cfg.Rules {
		cfg.Rules[name].Enabled = false
	}
	cfg.Rules[ruleName].Enabled = true
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

func BenchmarkIsolated_FinalBlankLine(b *testing.B)       { benchSingleRule(b, "final-blank-line", 1000) }
func BenchmarkIsolated_UnclosedCodeBlock(b *testing.B)    { benchSingleRule(b, "unclosed-code-block", 1000) }
func BenchmarkIsolated_EmptyAltText(b *testing.B)         { benchSingleRule(b, "empty-alt-text", 1000) }
func BenchmarkIsolated_HeadingLevel(b *testing.B)         { benchSingleRule(b, "heading-level", 1000) }
func BenchmarkIsolated_FencedCodeLanguage(b *testing.B)   { benchSingleRule(b, "fenced-code-language", 1000) }
func BenchmarkIsolated_DuplicateHeading(b *testing.B)     { benchSingleRule(b, "duplicate-heading", 1000) }
func BenchmarkIsolated_NoMultipleBlankLines(b *testing.B) { benchSingleRule(b, "no-multiple-blank-lines", 1000) }
func BenchmarkIsolated_NoSetextHeadings(b *testing.B)     { benchSingleRule(b, "no-setext-headings", 1000) }
func BenchmarkIsolated_SingleH1(b *testing.B)             { benchSingleRule(b, "single-h1", 1000) }
func BenchmarkIsolated_BlanksAroundHeadings(b *testing.B) { benchSingleRule(b, "blanks-around-headings", 1000) }
func BenchmarkIsolated_NoBareURLs(b *testing.B)           { benchSingleRule(b, "no-bare-urls", 1000) }
func BenchmarkIsolated_NoEmptyLinks(b *testing.B)         { benchSingleRule(b, "no-empty-links", 1000) }
func BenchmarkIsolated_NoTrailingSpaces(b *testing.B)     { benchSingleRule(b, "no-trailing-spaces", 1000) }

// --- Dummy rule tests: does adding ANY no-op rule cause regression? ---

// BenchmarkWith12PlusNoop runs 12 real rules + noop (no-trailing-spaces disabled).
// Tests if a 13th call site in collectErrors causes regression on its own.
func BenchmarkWith12PlusNoop(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	cfg.Rules["no-trailing-spaces"].Enabled = false
	cfg.Rules["noop"] = &config.RuleConfig{Enabled: true, Severity: config.SeverityError, Options: map[string]interface{}{}}
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// BenchmarkWith13PlusNoop runs 13 real rules + noop (14 total).
// Tests if the regression scales with rule count.
func BenchmarkWith13PlusNoop(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	cfg.Rules["noop"] = &config.RuleConfig{Enabled: true, Severity: config.SeverityError, Options: map[string]interface{}{}}
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}
