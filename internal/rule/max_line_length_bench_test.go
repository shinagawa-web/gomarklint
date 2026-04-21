package rule

import (
	"fmt"
	"strings"
	"testing"
)

func generateMarkdownWithLongLines(sections int) string {
	var sb strings.Builder
	for i := 1; i <= sections; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("Short line.\n\n")
		// one long line per section
		sb.WriteString(strings.Repeat("a", 100) + "\n\n")
	}
	return sb.String()
}

func BenchmarkCheckMaxLineLength(b *testing.B) {
	content := generateMarkdownWithLongLines(1000)
	lines := strings.Split(content, "\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckMaxLineLength("bench.md", lines, 0, 80)
	}
}
