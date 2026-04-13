package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownWithTrailingSpaces generates markdown with a mix of clean
// lines and lines that have trailing whitespace.
func generateMarkdownWithTrailingSpaces(sections int) string {
	var sb strings.Builder

	for i := 1; i <= sections; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("This is a clean line.\n")

		// Every 3rd line has trailing spaces
		if i%3 == 0 {
			sb.WriteString("This line has trailing spaces.   \n")
		} else {
			sb.WriteString("This line is also clean.\n")
		}

		sb.WriteString("\n")

		// Every 5th section has a fenced code block with trailing spaces inside
		if i%5 == 0 {
			sb.WriteString("```go\n")
			sb.WriteString("func example() {}   \n")
			sb.WriteString("```\n\n")
		}
	}

	return sb.String()
}

func BenchmarkCheckNoTrailingSpaces(b *testing.B) {
	content := generateMarkdownWithTrailingSpaces(1000)
	lines := strings.Split(content, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckNoTrailingSpaces("test.md", content, lines, 0)
	}
}
