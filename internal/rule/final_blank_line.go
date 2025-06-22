package rule

import (
	"strings"
)

func CheckFinalBlankLine(content string) []LintError {
	var errs []LintError

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || lines[len(lines)-1] != "" {
		errs = append(errs, LintError{
			Line:    len(lines),
			Message: "Missing final blank line",
		})
	}

	return errs
}
