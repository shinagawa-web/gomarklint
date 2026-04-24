package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
)

// benchmarkConfig returns the linter configuration used for all CI benchmarks.
// external-link is excluded (network cost).
// heading-level uses minLevel:1 to match the H1 intro; PR-7 will raise it to 2.
// max-line-length is enabled at 120 to exercise the rule without violations.
func benchmarkConfig() config.Config {
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false
	cfg.Rules["heading-level"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"minLevel": float64(1)},
	}
	cfg.Rules["max-line-length"] = &config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityError,
		Options:  map[string]interface{}{"lineLength": float64(120)},
	}
	return cfg
}

func writeIntro(sb *strings.Builder) {
	sb.WriteString("# Main Title\n\n")
	sb.WriteString("This is the introduction to the document.\n\n")
}

func writeHeading(sb *strings.Builder, i int) {
	fmt.Fprintf(sb, "## Section %d\n\n", i)
}

func writeParagraph(sb *strings.Builder) {
	sb.WriteString("This section contains important information. ")
	sb.WriteString("Here are some details that you should know about.\n\n")
}

func writeList(sb *strings.Builder) {
	sb.WriteString("Key points:\n\n")
	sb.WriteString("- First important point\n")
	sb.WriteString("- Second critical detail\n")
	sb.WriteString("- Third consideration\n\n")
}

func writeCodeBlock(sb *strings.Builder) {
	sb.WriteString("```go\n")
	sb.WriteString("func example() error {\n")
	sb.WriteString("    return nil\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")
}

func writeLinks(sb *strings.Builder, i int) {
	sb.WriteString("Useful resources:\n\n")
	fmt.Fprintf(sb, "- [Documentation](https://example.com/docs/%d)\n", i)
	fmt.Fprintf(sb, "- [GitHub](https://github.com/project/%d)\n", i)
	sb.WriteString("\n")
}

func writeImage(sb *strings.Builder, i int) {
	fmt.Fprintf(sb, "![Diagram %d](diagram%d.png)\n\n", i, i)
}

func writeSubsection(sb *strings.Builder, i int) {
	fmt.Fprintf(sb, "### Subsection %d.1\n\n", i)
	sb.WriteString("More detailed information goes here.\n\n")
}

// generateComplexMarkdown generates a realistic markdown file with mixed content.
func generateComplexMarkdown(sections int) string {
	var sb strings.Builder

	writeIntro(&sb)

	for i := 1; i <= sections; i++ {
		writeHeading(&sb, i)
		writeParagraph(&sb)
		writeList(&sb)

		if i%2 == 0 {
			writeCodeBlock(&sb)
		}

		if i%3 == 0 {
			writeLinks(&sb, i)
		}

		if i%4 == 0 {
			writeImage(&sb, i)
		}

		writeSubsection(&sb, i)
	}

	// Each section ends with \n\n; strip the trailing extra newline so the
	// document ends with exactly one newline (satisfies final-blank-line) and
	// does not trigger no-multiple-blank-lines at EOF.
	result := sb.String()
	return result[:len(result)-1]
}

func BenchmarkFullLinting(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := benchmarkConfig()
	lint := linter.New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

func BenchmarkFullLinting_ExtraLarge(b *testing.B) {
	content := generateComplexMarkdown(5000)
	cfg := benchmarkConfig()
	lint := linter.New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}
