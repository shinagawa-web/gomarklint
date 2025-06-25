package rule

import (
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"strings"
)

func CheckFinalBlankLine(filename, content string) []LintError {
	body, offset := parser.StripFrontmatter(content)

	var errs []LintError

	lines := strings.Split(body, "\n")
	if len(lines) < 2 || lines[len(lines)-1] != "" {
		errs = append(errs, LintError{
			File:    filename,
			Line:    len(lines) + offset,
			Message: "Missing final blank line",
		})
	}

	return errs
}
