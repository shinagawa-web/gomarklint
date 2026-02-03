package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownWithCodeBlocks generates markdown content with code blocks.
func generateMarkdownWithCodeBlocks(blocks int) string {
	var sb strings.Builder
	languages := []string{"go", "python", "javascript", "bash"}

	for i := 1; i <= blocks; i++ {
		sb.WriteString(fmt.Sprintf("## Code Example %d\n\n", i))
		lang := languages[i%len(languages)]

		sb.WriteString(fmt.Sprintf("```%s\n", lang))
		sb.WriteString("func example() {\n")
		sb.WriteString("    return nil\n")
		sb.WriteString("}\n")
		sb.WriteString("```\n\n")
	}
	return sb.String()
}

func BenchmarkCheckCodeBlocks(b *testing.B) {
	content := generateMarkdownWithCodeBlocks(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckUnclosedCodeBlocks("test.md", content)
	}
}
