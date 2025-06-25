package rule

import (
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"strings"
)

func CheckUnclosedCodeBlocks(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)

	var errs []LintError
	lines := strings.Split(body, "\n")

	inCodeBlock := false
	var startLine int

	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				startLine = i + 1 + offset
			} else {
				inCodeBlock = false
			}
		}
	}

	if inCodeBlock {
		errs = append(errs, LintError{
			File:    filename,
			Line:    startLine,
			Message: "Unclosed code block",
		})
	}

	return errs
}
