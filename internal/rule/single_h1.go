package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func CheckSingleH1(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError
	foundFirst := false

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		line := ctx.Line(i)
		if firstNonSpaceByte(line) != '#' {
			continue
		}

		trimmed := strings.TrimSpace(line)

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
