package rule

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// generateMarkdownWithLinks generates markdown content with many external links.
func generateMarkdownWithLinks(blocks int) string {
	var sb strings.Builder
	domains := []string{"example.com", "github.com", "golang.org", "test.com"}

	for i := 1; i <= blocks; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("Here are some useful links:\n\n")

		for j := 0; j < 5; j++ {
			domain := domains[j%len(domains)]
			sb.WriteString(fmt.Sprintf("- [Link %d-%d](https://%s/page%d)\n", i, j, domain, j))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func BenchmarkCheckExternalLinks(b *testing.B) {
	content := generateMarkdownWithLinks(1000)
	lines := strings.Split(content, "\n")
	offset := 0
	urlCache := &sync.Map{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckExternalLinks("test.md", lines, offset, nil, 0, 0, urlCache)
	}
}
