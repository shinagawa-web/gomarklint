package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// emptyLinkDest reports whether dest is an "empty" link destination.
// Empty means literally empty, a lone fragment "#", or angle-bracket-wrapped
// empty "<>".
func emptyLinkDest(dest string) bool {
	return dest == "" || dest == "#" || dest == "<>"
}

// findEmptyLinks scans a single line (already stripped of inline code) for
// Markdown links or images whose destination is empty.
// It returns the raw matched text for each violation (e.g. "[text]()").
func findEmptyLinks(line string) []string {
	var results []string
	pos := 0
	for pos < len(line) {
		// Look for '](' which signals link/image destination.
		idx := strings.Index(line[pos:], "](")
		if idx == -1 {
			break
		}
		openParen := pos + idx + 2 // position right after '('

		// Find the matching closing ')'.
		closeParen := strings.Index(line[openParen:], ")")
		if closeParen == -1 {
			break
		}
		closeParen += openParen

		dest := strings.TrimSpace(line[openParen:closeParen])
		if emptyLinkDest(dest) {
			// Walk back to find the opening '[' (or '![').
			bracketStart := pos + idx
			for bracketStart > 0 && line[bracketStart-1] != '[' {
				bracketStart--
			}
			if bracketStart > 0 && line[bracketStart-1] == '[' {
				bracketStart--
				// Check for image prefix '!'.
				if bracketStart > 0 && line[bracketStart-1] == '!' {
					bracketStart--
				}
			}
			results = append(results, line[bracketStart:closeParen+1])
		}
		pos = closeParen + 1
	}
	return results
}

// CheckNoEmptyLinks flags Markdown links and images whose destination URL is
// empty, contains only "#", or is "<>".
// Links inside fenced code, indented code, HTML blocks, HTML comments, and inline
// code spans are ignored.
func CheckNoEmptyLinks(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		// Sanitized blanks inline code spans and inline comments, so links
		// living inside them are not seen here.
		line := ctx.Sanitized(i)
		if !strings.Contains(line, "](") {
			continue
		}

		for _, match := range findEmptyLinks(line) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("no-empty-links: link has empty destination: %s", match),
			})
		}
	}

	return errs
}
