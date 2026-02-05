package rule

import (
	"regexp"
)

// CheckEmptyAltText checks for image tags with empty alt text (![](...))
//
// Parameters:
//   - filename: the name of the file being checked
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError if any image has empty alt text.
func CheckEmptyAltText(filename string, lines []string, offset int) []LintError {
	var errs []LintError
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
