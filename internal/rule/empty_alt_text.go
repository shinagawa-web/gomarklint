package rule

import (
	"github.com/shinagawa-web/gomarklint/internal/parser"

	"regexp"
	"strings"
)

// CheckEmptyAltText checks for image tags with empty alt text (![](...))
//
// Parameters:
//   - filename: the name of the file being checked
//   - content: the raw Markdown content
//
// Returns:
//   - A slice of LintError if any image has empty alt text.
func CheckEmptyAltText(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)

	var errs []LintError
	lines := strings.Split(body, "\n")
	re := regexp.MustCompile(`!\[\s*\]\([^)]+\)`)

	for i, line := range lines {
		if re.MatchString(line) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: "image with empty alt text",
			})
		}
	}

	return errs
}
