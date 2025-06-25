package rule

import (
	"github.com/shinagawa-web/gomarklint/internal/parser"
	"strings"
)

// CheckUnclosedCodeBlocks detects any unclosed fenced code blocks (e.g., ```)
// in the Markdown content. It ensures that every opening fence has a corresponding closing fence.
//
// It also handles frontmatter by skipping it and adjusting line numbers.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - content: the raw Markdown content as a string
//
// Returns:
//   - A slice of LintError indicating the location of any unclosed code block.
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
