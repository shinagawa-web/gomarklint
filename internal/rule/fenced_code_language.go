package rule

import (
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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
