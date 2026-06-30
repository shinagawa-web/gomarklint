package rule

import (
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
