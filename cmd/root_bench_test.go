package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
)

// generateComplexMarkdown generates a realistic markdown file with mixed content,
// including gomarklint-disable directives scattered throughout so the disable-comment
// parse and filter path is exercised on branches that support it.
func generateComplexMarkdown(sections int) string {
	var sb strings.Builder

	sb.WriteString("# Main Title\n\n")
	sb.WriteString("This is the introduction to the document.\n\n")

	for i := 1; i <= sections; i++ {
		// Heading
		fmt.Fprintf(&sb, "## Section %d\n\n", i)

		// Paragraph
		sb.WriteString("This section contains important information. ")
		sb.WriteString("Here are some details that you should know about.\n\n")

		// List
		sb.WriteString("Key points:\n\n")
		sb.WriteString("- First important point\n")
		sb.WriteString("- Second critical detail\n")
		sb.WriteString("- Third consideration\n\n")

		// Code block
		if i%2 == 0 {
			sb.WriteString("```go\n")
			sb.WriteString("func example() error {\n")
			sb.WriteString("    return nil\n")
			sb.WriteString("}\n")
			sb.WriteString("```\n\n")
		}

		// Links
		if i%3 == 0 {
			sb.WriteString("Useful resources:\n\n")
			fmt.Fprintf(&sb, "- [Documentation](https://example.com/docs/%d)\n", i)
			fmt.Fprintf(&sb, "- [GitHub](https://github.com/project/%d)\n", i)
			sb.WriteString("\n")
		}

		// Image
		if i%4 == 0 {
			fmt.Fprintf(&sb, "![Diagram %d](diagram%d.png)\n\n", i, i)
		}

		// Every 5th section: block disable/enable
		if i%5 == 0 {
			sb.WriteString("<!-- gomarklint-disable no-bare-urls -->\n")
			fmt.Fprintf(&sb, "https://suppressed-%d.example.com\n", i)
			sb.WriteString("<!-- gomarklint-enable no-bare-urls -->\n\n")
		}

		// Every 7th section: disable-line
		if i%7 == 0 {
			fmt.Fprintf(&sb, "https://inline-%d.example.com <!-- gomarklint-disable-line no-bare-urls -->\n\n", i)
		}

		// Every 11th section: disable-next-line
		if i%11 == 0 {
			sb.WriteString("<!-- gomarklint-disable-next-line no-bare-urls -->\n")
			fmt.Fprintf(&sb, "https://nextline-%d.example.com\n\n", i)
		}

		// Subsection
		fmt.Fprintf(&sb, "### Subsection %d.1\n\n", i)
		sb.WriteString("More detailed information goes here.\n\n")
	}

	return sb.String()
}

func BenchmarkFullLinting(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false

	lint := linter.New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}

func BenchmarkFullLinting_ExtraLarge(b *testing.B) {
	content := generateComplexMarkdown(5000)
	cfg := config.Default()
	cfg.Rules["external-link"].Enabled = false

	lint := linter.New(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = lint.LintContent("benchmark.md", content)
	}
}
