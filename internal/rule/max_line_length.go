package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func isBareURLLine(trimmed string) bool {
	s := trimmed
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
		s = s[1 : len(s)-1]
	}
	var schemeLen int
	if strings.HasPrefix(s, "https://") {
		schemeLen = len("https://")
	} else if strings.HasPrefix(s, "http://") {
		schemeLen = len("http://")
	} else {
		return false
	}
	// The body after the scheme must not contain whitespace — ensuring
	// the entire line is exactly one URL token with no surrounding text.
	return !strings.ContainsAny(s[schemeLen:], " \t")
}

func CheckMaxLineLength(filename string, ctx *preprocess.Context, offset int, lineLength int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if ctx.InFencedCode(i) {
			continue
		}

		line := ctx.Line(i)
		if len(line) <= lineLength {
			continue
		}

		trimmed := strings.TrimSpace(line)
		first := firstNonSpaceByte(line)
		if (first == '#' && isATXHeading(trimmed)) || isBareURLLine(trimmed) {
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: fmt.Sprintf("max-line-length: line exceeds %d bytes (%d)", lineLength, len(line)),
		})
	}

	return errs
}
