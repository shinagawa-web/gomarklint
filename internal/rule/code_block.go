package rule

import (
	"strings"
)

// CheckUnclosedCodeBlocks detects any unclosed fenced code blocks (e.g., ```)
// in the Markdown content. It ensures that every opening fence has a corresponding closing fence.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - lines: the Markdown content split into lines (with frontmatter already removed)
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError indicating the location of any unclosed code block.
func CheckUnclosedCodeBlocks(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	_, unclosed := GetCodeBlockLineRanges(lines)

	for _, start := range unclosed {
		errs = append(errs, LintError{
			File:    filename,
			Line:    start + offset + 1,
			Message: "Unclosed code block",
		})
	}

	return errs
}

func GetCodeBlockLineRanges(lines []string) (closed [][2]int, unclosed []int) {
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
