package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// CheckNoHardTabs flags hard tab characters (\t) outside fenced code blocks
// and inline code spans. Each tab is reported as a separate violation.
//
// This is a markdownlint divergence (#337 Section B): only fenced code is
// skipped (tabs in indented code are still reported), and inline code spans are
// stripped. It therefore skips just preprocess.InFencedCode and keeps its own
// inline-code stripping rather than using the shared block-context helper or the
// sanitized view (which would also blank inline HTML comments).
func CheckNoHardTabs(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if ctx.InFencedCode(i) {
			continue
		}

		line := ctx.Line(i)
		if !strings.ContainsRune(line, '\t') {
			continue
		}

		scanned := line
		if strings.ContainsRune(line, '`') {
			scanned = stripInlineCode(line)
		}

		col := 0
		for _, ch := range scanned {
			col++
			if ch == '\t' {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: fmt.Sprintf("no-hard-tabs: hard tab character found at column %d", col),
				})
			}
		}
	}

	return errs
}
