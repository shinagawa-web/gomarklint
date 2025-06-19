package rule

import (
	"fmt"
	"regexp"
	"strings"
)

type LintError struct {
	Line    int
	Message string
}

func CheckHeadingLevels(content string) []LintError {
	lines := strings.Split(content, "\n")
	var errs []LintError

	prevLevel := 0
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+`)

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			currentLevel := len(matches[1])
			if prevLevel != 0 && currentLevel > prevLevel+1 {
				errs = append(errs, LintError{
					Line:    i + 1,
					Message: fmt.Sprintf("Heading level jumped from %d to %d", prevLevel, currentLevel),
				})
			}
			prevLevel = currentLevel
		}
	}

	return errs
}
