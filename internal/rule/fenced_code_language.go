package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckFencedCodeLanguage checks that all fenced code blocks specify a language identifier.
// Fenced code blocks opened with ``` or ~~~ without a language tag are flagged (MD040).
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError, one per opening fence that is missing a language identifier.
//
// Fence openers inside indented code, HTML blocks, and HTML comments are not real
// fences and are excluded by the scanner, so they are never examined.
func CheckFencedCodeLanguage(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for _, span := range ctx.FenceSpans() {
		trimmed := strings.TrimSpace(ctx.Line(span.Start))
		marker := openingFenceMarker(trimmed)
		if strings.TrimSpace(trimmed[len(marker):]) == "" {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + span.Start + 1,
				Message: "Fenced code block must have a language identifier",
			})
		}
	}

	return errs
}
