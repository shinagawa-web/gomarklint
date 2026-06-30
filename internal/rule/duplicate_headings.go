package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckDuplicateHeadings detects duplicate headings in the given Markdown content.
// It treats headings as duplicates if their normalized text matches, regardless of level,
// case differences, or trailing spaces (including full-width spaces).
//
// This check helps enforce clear and non-redundant structure in documents,
// which improves readability and avoids confusion.
//
// Parameters:
//   - filename: the name of the file being checked (used in error reporting)
//   - ctx: the shared per-line context produced by preprocess.Scan
//   - offset: the line number offset due to frontmatter removal
//
// Returns:
//   - A slice of LintError entries for each detected duplicate heading (excluding the first occurrence).
//
// Headings inside fenced code, indented code, HTML blocks, and HTML comments are
// ignored.
func CheckDuplicateHeadings(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	seen := make(map[string]struct{}, ctx.Len()/10)

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		line := ctx.Line(i)
		if firstNonSpaceByte(line) != '#' {
			continue
		}

		trimmed := strings.TrimSpace(line)

		if !isATXHeading(trimmed) {
			continue
		}

		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		normalized := strings.ToLower(heading)

		if _, ok := seen[normalized]; ok {
			errs = append(errs, LintError{
				File:    filename,
				Line:    i + 1 + offset,
				Message: fmt.Sprintf("duplicate heading: %q", normalized),
			})
		} else {
			seen[normalized] = struct{}{}
		}
	}

	return errs
}
