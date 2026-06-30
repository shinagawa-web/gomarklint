package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckNoMultipleBlankLines checks whether the Markdown content contains
// multiple consecutive blank lines.
//
// This rule helps maintain consistency and readability by ensuring
// no more than one consecutive blank line appears in the document.
//
// It ignores multiple consecutive newlines in code blocks.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError with entries for each occurrence of multiple consecutive blank lines.
//
// Blank lines inside any block context the scanner identifies are not counted —
// fenced code, but also blank lines inside type 1–5 HTML blocks (e.g. <pre>) and
// multi-line HTML comments. Blank lines between indented-code paragraphs are not
// classified as indented by the scanner, so that case remains a known limitation.
func CheckNoMultipleBlankLines(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	consecutiveBlankCount := 0

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			consecutiveBlankCount = 0
			continue
		}
		if strings.TrimSpace(ctx.Line(i)) == "" {
			consecutiveBlankCount++
			if consecutiveBlankCount > 1 {
				errs = append(errs, LintError{
					File:    filename,
					Line:    i + 1 + offset,
					Message: "Multiple consecutive blank lines",
				})
			}
		} else {
			consecutiveBlankCount = 0
		}
	}

	return errs
}
