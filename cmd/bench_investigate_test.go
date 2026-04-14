package cmd

import (
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
)

func makeConfig(noopCount int, enableTrailing bool) config.Config {
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	if !enableTrailing {
		cfg.Rules["no-trailing-spaces"].Enabled = false
	}
	noopNames := []string{"noop1", "noop2", "noop3", "noop4", "noop5", "noop6", "noop7", "noop8", "noop9", "noop10", "noop11", "noop12", "noop13"}
	for i := 0; i < noopCount && i < len(noopNames); i++ {
		cfg.Rules[noopNames[i]] = &config.RuleConfig{Enabled: true, Severity: config.SeverityError, Options: map[string]interface{}{}}
	}
	return cfg
}

// --- Test A: How does if-block count affect performance? ---
// All use the SAME compiled binary. Vary enabled noop rules.
// Total if-blocks in source: 13 real + 13 noop = 26
// Enabled rules: 12 base + no-trailing-spaces + N noops

// 12 rules (no-trailing-spaces OFF, 0 noops)
func BenchmarkBlocks12(b *testing.B) {
	content := generateComplexMarkdown(1000)
	lint := linter.New(makeConfig(0, false))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// 13 rules (no-trailing-spaces ON, 0 noops)
func BenchmarkBlocks13(b *testing.B) {
	content := generateComplexMarkdown(1000)
	lint := linter.New(makeConfig(0, true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// 15 rules (13 + 2 noops)
func BenchmarkBlocks15(b *testing.B) {
	content := generateComplexMarkdown(1000)
	lint := linter.New(makeConfig(2, true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// 20 rules (13 + 7 noops)
func BenchmarkBlocks20(b *testing.B) {
	content := generateComplexMarkdown(1000)
	lint := linter.New(makeConfig(7, true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// 25 rules (13 + 12 noops)
func BenchmarkBlocks25(b *testing.B) {
	content := generateComplexMarkdown(1000)
	lint := linter.New(makeConfig(12, true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

// --- Test B: Does splitting collectErrors into smaller functions help? ---

// BenchmarkSplitFunction uses collectErrorsSplit (two groups of ~6-7 rules)
func BenchmarkSplitFunction(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContentSplit("benchmark.md", content)
	}
}

// BenchmarkSingleFunction is the same rules but via the original single collectErrors
func BenchmarkSingleFunction(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	lint := linter.New(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}
