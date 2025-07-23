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
	_, unclosed := GetCodeBlockLineRanges(body)

	for _, start := range unclosed {
		errs = append(errs, LintError{
			File:    filename,
			Line:    start + offset + 1,
			Message: "Unclosed code block",
		})
	}

	return errs
}

func GetCodeBlockLineRanges(content string) (closed [][2]int, unclosed []int) {
	lines := strings.Split(content, "\n")
	inBlock := false
	var start int
	var closedRanges [][2]int
	var unclosedStarts []int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				closedRanges = append(closedRanges, [2]int{start + 1, i + 1})
				inBlock = false
			} else {
				start = i
				inBlock = true
			}
		}
	}

	if inBlock {
		unclosedStarts = append(unclosedStarts, start)
	}

	return closedRanges, unclosedStarts
}
