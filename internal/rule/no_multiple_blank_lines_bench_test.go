package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownWithBlankLines generates markdown with various blank line patterns.
func generateMarkdownWithBlankLines(blocks int) string {
	var sb strings.Builder
	
	for i := 1; i <= blocks; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("First paragraph.\n")
		
		// Add varying numbers of blank lines
		blankLines := (i % 3) + 1
		for j := 0; j < blankLines; j++ {
			sb.WriteString("\n")
		}
		
		sb.WriteString("Second paragraph.\n\n")
	}
	return sb.String()
}

func BenchmarkCheckNoMultipleBlankLines(b *testing.B) {
	content := generateMarkdownWithBlankLines(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckNoMultipleBlankLines("test.md", content)
	}
}
