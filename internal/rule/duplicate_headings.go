package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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
