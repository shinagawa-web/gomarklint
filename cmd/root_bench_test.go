package cmd

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/config"
)

// generateComplexMarkdown generates a realistic markdown file with mixed content.
func generateComplexMarkdown(sections int) string {
	var sb strings.Builder

	sb.WriteString("# Main Title\n\n")
	sb.WriteString("This is the introduction to the document.\n\n")

	for i := 1; i <= sections; i++ {
		// Heading
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))

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
			sb.WriteString(fmt.Sprintf("- [Documentation](https://example.com/docs/%d)\n", i))
			sb.WriteString(fmt.Sprintf("- [GitHub](https://github.com/project/%d)\n", i))
			sb.WriteString("\n")
		}

		// Image
		if i%4 == 0 {
			sb.WriteString(fmt.Sprintf("![Diagram %d](diagram%d.png)\n\n", i, i))
		}

		// Subsection
		sb.WriteString(fmt.Sprintf("### Subsection %d.1\n\n", i))
		sb.WriteString("More detailed information goes here.\n\n")
	}

	return sb.String()
}

func BenchmarkFullLinting(b *testing.B) {
	content := generateComplexMarkdown(1000)
	cfg := config.Default()
	cfg.EnableLinkCheck = false
	urlCache := &sync.Map{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = collectErrors("benchmark.md", content, cfg, nil, urlCache)
	}
}

func BenchmarkFullLinting_ExtraLarge(b *testing.B) {
	content := generateComplexMarkdown(5000)
	cfg := config.Default()
	cfg.EnableLinkCheck = false
	urlCache := &sync.Map{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = collectErrors("benchmark.md", content, cfg, nil, urlCache)
	}
}
