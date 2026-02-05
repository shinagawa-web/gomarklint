package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownWithImages generates markdown content with image tags.
func generateMarkdownWithImages(blocks int) string {
	var sb strings.Builder

	for i := 1; i <= blocks; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))

		// Mix of images with and without alt text
		if i%2 == 0 {
			sb.WriteString(fmt.Sprintf("![Image %d](image%d.png)\n\n", i, i))
		} else {
			sb.WriteString(fmt.Sprintf("![](image%d.png)\n\n", i))
		}
	}
	return sb.String()
}

func BenchmarkCheckEmptyAltText(b *testing.B) {
	content := generateMarkdownWithImages(1000)
	lines := strings.Split(content, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckEmptyAltText("test.md", lines, 0)
	}
}
