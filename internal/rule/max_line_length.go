package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// isBareURLLine reports whether the trimmed line consists solely of a single
// URL token (http:// or https://), possibly wrapped in angle brackets.
// Lines like "https://example.com extra text" are NOT exempt.
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

// CheckMaxLineLength flags lines whose byte length exceeds lineLength.
// Lines inside fenced code blocks, ATX heading lines, and lines that consist
// solely of a URL are exempt.
//
// This is a markdownlint divergence (#337 Section B): only fenced code is
// skipped. Lines inside indented code and HTML blocks are deliberately still
// checked, so the rule skips just preprocess.InFencedCode rather than every
// block context.
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
