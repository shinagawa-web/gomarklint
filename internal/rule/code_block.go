package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckUnclosedCodeBlocks detects any unclosed fenced code blocks (e.g., ```)
// in the Markdown content. It ensures that every opening fence has a corresponding closing fence.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError indicating the location of any unclosed code block.
//
// Because fence detection comes from the shared scanner, fence markers inside
// indented code or HTML blocks can no longer be mispaired into phantom unclosed
// blocks (audit #337 cascade).
func CheckUnclosedCodeBlocks(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for _, span := range ctx.FenceSpans() {
		if span.End == -1 {
			errs = append(errs, LintError{
				File:    filename,
				Line:    span.Start + offset + 1,
				Message: "Unclosed code block",
			})
		}
	}

	return errs
}

// GetCodeBlockLineRanges scans lines and returns the 1-based line ranges of all
// closed fenced code blocks. It is retained for external-link, the last rule not
// yet migrated to the preprocess context (#337 Phase 3); unclosed-code-block now
// derives unclosed blocks from preprocess.Context.FenceSpans instead.
//
// Optimization: a byte-level prefilter (firstNonSpaceByte) avoids calling
// strings.TrimSpace on the ~95% of lines that cannot be fence openers or
// closers. Inside a block, only a line whose first non-space byte matches the
// opening fence character can close the block.
func GetCodeBlockLineRanges(lines []string) [][2]int {
	inBlock := false
	var start int
	var fenceMarker string
	var closedRanges [][2]int

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

	return closedRanges
}
