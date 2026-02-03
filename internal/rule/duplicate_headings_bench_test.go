package rule

import (
	"fmt"
	"strings"
	"testing"
)

// generateMarkdownWithDuplicates generates markdown with some duplicate headings.
func generateMarkdownWithDuplicates(blocks int) string {
	var sb strings.Builder
	headingNames := []string{"Introduction", "Overview", "Details", "Summary", "Conclusion"}
	
	for i := 1; i <= blocks; i++ {
		heading := headingNames[i%len(headingNames)]
		sb.WriteString(fmt.Sprintf("## %s\n\n", heading))
		sb.WriteString("Some content goes here.\n\n")
	}
	return sb.String()
}

func BenchmarkCheckDuplicateHeadings(b *testing.B) {
	content := generateMarkdownWithDuplicates(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckDuplicateHeadings("test.md", content)
	}
}
