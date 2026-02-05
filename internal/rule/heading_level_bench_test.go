package rule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/config"
)

// generateMarkdownWithHeadings generates markdown content with various heading levels.
func generateMarkdownWithHeadings(blocks int) string {
	var sb strings.Builder
	for i := 1; i <= blocks; i++ {
		level := (i % 4) + 1 // Cycle through levels 1-4
		sb.WriteString(fmt.Sprintf("%s Heading %d\n\n", strings.Repeat("#", level), i))
		sb.WriteString("This is a paragraph with some content. ")
		sb.WriteString("It contains multiple sentences to make it more realistic.\n\n")
	}
	return sb.String()
}

func BenchmarkCheckHeadingLevel(b *testing.B) {
	content := generateMarkdownWithHeadings(1000)
	cfg := config.Default()
	lines := strings.Split(content, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckHeadingLevels("test.md", lines, 0, cfg.MinHeadingLevel)
	}
}
