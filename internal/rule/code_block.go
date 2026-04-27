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

// GetCodeBlockLineRanges scans lines and returns the 1-based line ranges of all
// closed fenced code blocks and the 0-based start lines of any unclosed ones.
//
// Optimisation: a byte-level prefilter (firstNonSpaceByte) avoids calling
// strings.TrimSpace on the ~95% of lines that cannot be fence openers or
// closers. Inside a block, only a line whose first non-space byte matches the
// opening fence character can close the block.
func GetCodeBlockLineRanges(lines []string) (closed [][2]int, unclosed []int) {
	inBlock := false
	var start int
	var fenceMarker string
	var closedRanges [][2]int
	var unclosedStarts []int

	for i, line := range lines {
		first := firstNonSpaceByte(line)
		if inBlock {
			// Only a line whose first non-space byte matches the opener's fence
			// character can close the block; all others are content.
			if first != fenceMarker[0] {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if IsClosingFence(trimmed, fenceMarker) {
				closedRanges = append(closedRanges, [2]int{start + 1, i + 1})
				inBlock = false
				fenceMarker = ""
			}
			continue
		}
		// Outside a block, only backtick or tilde lines can open a fence.
		if first != '`' && first != '~' {
			continue
		}
		trimmed := strings.TrimSpace(line)
		marker := openingFenceMarker(trimmed)
		if marker != "" {
			start = i
			fenceMarker = marker
			inBlock = true
		}
	}

	if inBlock {
		unclosedStarts = append(unclosedStarts, start)
	}

	return closedRanges, unclosedStarts
}
