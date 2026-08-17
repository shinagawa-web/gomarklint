package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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
