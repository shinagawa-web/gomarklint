package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownContent generates generic markdown content.
func generateMarkdownContent(lines int) string {
	var sb strings.Builder

	for i := 1; i <= lines; i++ {
		if i%10 == 1 {
			sb.WriteString(fmt.Sprintf("# Heading %d\n", i))
		} else {
			sb.WriteString(fmt.Sprintf("Line %d with some text content.\n", i))
		}
	}
	return sb.String()
}

func BenchmarkCheckFinalBlankLine(b *testing.B) {
	content := generateMarkdownContent(5000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckFinalBlankLine("test.md", content)
	}
}
