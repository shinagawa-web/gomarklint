package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckSingleH1 flags every ATX-style H1 heading (`# ...`) after the first one in the file.
// H1 headings inside fenced code, indented code, HTML blocks, and HTML comments
// are ignored.
func CheckSingleH1(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	foundFirst := false

	for i := 0; i < ctx.Len(); i++ {
		// Skip lines the scanner classified as code or HTML: an H1-looking line
		// there is not a real heading.
		if inBlockContext(ctx, i) {
			continue
		}

		// Byte-level prefilter: skip lines whose first non-ASCII-space byte
		// cannot start an H1 heading, avoiding strings.TrimSpace on the vast
		// majority of lines (paragraphs, list items, blank lines, etc.).
		line := ctx.Line(i)
		if firstNonSpaceByte(line) != '#' {
			continue
		}

		// The prefilter guarantees the first non-space byte is '#', so the
		// trimmed line begins with '#'.
		trimmed := strings.TrimSpace(line)

		// Must be "# ..." (H1 with space) or bare "#" (also H1).
		if len(trimmed) >= 2 && trimmed[1] != ' ' {
			continue
		}

		if !foundFirst {
			foundFirst = true
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: "Multiple H1 headings found; only one H1 is allowed per file",
		})
	}

	return errs
}
