package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func matchDoubleDelim(line, delim string) (string, bool) {
	if len(line) <= len(delim)*2 {
		return "", false
	}
	if !strings.HasPrefix(line, delim) || !strings.HasSuffix(line, delim) {
		return "", false
	}
	inner := line[len(delim) : len(line)-len(delim)]
	if strings.Contains(inner, delim) {
		return "", false
	}
	return inner, true
}

func matchSingleDelim(line string, ch byte) (string, bool) {
	double := string([]byte{ch, ch})
	if len(line) <= 2 || line[0] != ch || line[len(line)-1] != ch {
		return "", false
	}
	if strings.HasPrefix(line, double) {
		return "", false
	}
	inner := line[1 : len(line)-1]
	if strings.ContainsRune(inner, rune(ch)) {
		return "", false
	}
	return inner, true
}

func emphasisContent(line string) (string, bool) {
	if inner, ok := matchDoubleDelim(line, "**"); ok {
		return inner, true
	}
	if inner, ok := matchDoubleDelim(line, "__"); ok {
		return inner, true
	}
	if inner, ok := matchSingleDelim(line, '*'); ok {
		return inner, true
	}
	if inner, ok := matchSingleDelim(line, '_'); ok {
		return inner, true
	}
	return "", false
}

var punctuationChars = map[rune]struct{}{
	'.': {}, ',': {}, ';': {}, ':': {}, '!': {}, '?': {}, // ASCII
	'。': {}, '、': {}, '；': {}, '：': {}, '！': {}, '？': {}, // Full-width / CJK
}

func endsWithPunctuation(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	_, ok := punctuationChars[runes[len(runes)-1]]
	return ok
}

func CheckNoEmphasisAsHeading(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		line := ctx.Line(i)
		first := firstNonSpaceByte(line)
		if first != '*' && first != '_' {
			continue
		}

		trimmed := strings.TrimSpace(line)
		inner, ok := emphasisContent(trimmed)
		if !ok {
			continue
		}
		// Trim nested emphasis delimiters from inner before the punctuation
		// check so that e.g. "***Note:***" (inner="*Note:*") is correctly
		// recognized as ending with ':' rather than '*'.
		if endsWithPunctuation(strings.TrimRight(inner, "*_")) {
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: fmt.Sprintf("no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: %s", trimmed),
		})
	}

	return errs
}
