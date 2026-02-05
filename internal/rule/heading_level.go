package rule

import (
	"fmt"
	"regexp"
)

type LintError struct {
	File    string
	Line    int
	Message string
}

// CheckHeadingLevels analyzes the heading structure of the given Markdown content
// and reports any issues such as the first heading not starting at the specified minimum level
// or heading levels that jump more than one level (e.g., from ## to ####).
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//   - minLevel: the expected minimum level for the first heading (e.g., 2 for ##)
//
// Returns:
//   - A slice of LintError containing the line number and description of each detected issue.
func CheckHeadingLevels(filename string, lines []string, offset int, minLevel int) []LintError {
	var errs []LintError

	prevLevel := 0
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+`)

	codeBlockRanges, _ := GetCodeBlockLineRanges(lines)

	for i, line := range lines {
		if isInCodeBlock(i+1, codeBlockRanges) {
			continue
		}
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			currentLevel := len(matches[1])

			if prevLevel == 0 {
				if currentLevel != minLevel {
					errs = append(errs, LintError{
						File:    filename,
						Line:    i + 1 + offset,
						Message: fmt.Sprintf("First heading should be level %d (found level %d)", minLevel, currentLevel),
					})
				}
			} else if currentLevel > prevLevel+1 {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: fmt.Sprintf("Heading level jumped from %d to %d", prevLevel, currentLevel),
				})
			}
			prevLevel = currentLevel
		}
	}

	return errs
}
