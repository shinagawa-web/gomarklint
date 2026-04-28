package rule

import (
	"regexp"
	"strings"
)

var emptyAltTextRe = regexp.MustCompile(`!\[\s*\]\([^)]+\)`)

// CheckEmptyAltText reports lint errors for images with empty alt text.
func CheckEmptyAltText(filename string, lines []string, offset int) []LintError {
	var errs []LintError

	for i, line := range lines {
		if !strings.Contains(line, "![") {
			continue
		}
		if emptyAltTextRe.MatchString(line) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: "image with empty alt text",
			})
		}
	}

	return errs
}
