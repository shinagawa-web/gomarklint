package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// isATXHeading reports whether s is an ATX-style heading (levels 1–6).
func isATXHeading(s string) bool {
	level := 0
	for level < len(s) && s[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return false
	}
	return level == len(s) || s[level] == ' ' || s[level] == '\t'
}

// CheckBlanksAroundHeadings flags ATX-style headings that are not preceded or
// followed by a blank line. Headings at the start or end of the file are exempt
// from the respective check. Headings inside fenced code, indented code, HTML
// blocks, and HTML comments are ignored.
func CheckBlanksAroundHeadings(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	// prevBlank tracks whether the previous line was blank so we avoid a
	// second TrimSpace call on lines[i-1] for every heading encountered.
	// Initialized to true so the first line of the file is exempt.
	prevBlank := true
	prevWasHeading := false

	for i := 0; i < ctx.Len(); i++ {
		trimmed := strings.TrimSpace(ctx.Line(i))
		isBlank := trimmed == ""

		// Check "followed by blank" before the block skip so a heading
		// immediately followed by a code/HTML block opener is still flagged.
		if prevWasHeading && !isBlank {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i,
				Message: "blanks-around-headings: heading must be followed by a blank line",
			})
		}

		if inBlockContext(ctx, i) {
			prevBlank = false
			prevWasHeading = false
			continue
		}

		if isATXHeading(trimmed) {
			if i > 0 && !prevBlank {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "blanks-around-headings: heading must be preceded by a blank line",
				})
			}
			prevWasHeading = true
		} else {
			prevWasHeading = false
		}

		prevBlank = isBlank
	}

	return errs
}
