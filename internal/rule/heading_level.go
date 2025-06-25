package rule

import (
	"fmt"
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"regexp"
	"strings"
)

type LintError struct {
	File    string
	Line    int
	Message string
}

func CheckHeadingLevels(filename, content string, minLevel int) []LintError {
	body, offset := parser.StripFrontmatter(content)
	lines := strings.Split(body, "\n")
	var errs []LintError

	prevLevel := 0
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+`)

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			currentLevel := len(matches[1])

			if prevLevel == 0 {
				if currentLevel != minLevel {
					errs = append(errs, LintError{
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
