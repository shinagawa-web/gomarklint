package rule

import (
	"strings"
)

func CheckUnclosedCodeBlocks(content string) []LintError {
	var errs []LintError
	lines := strings.Split(content, "\n")

	inCodeBlock := false
	var startLine int

	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				startLine = i + 1
			} else {
				inCodeBlock = false
			}
		}
	}

	if inCodeBlock {
		errs = append(errs, LintError{
			Line:    startLine,
			Message: "Unclosed code block",
		})
	}

	return errs
}
