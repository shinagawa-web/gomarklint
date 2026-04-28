package rule

import (
	"fmt"
	"strings"
)

// matchDoubleDelim returns the inner text if line is entirely wrapped in a
// two-character delimiter (e.g. "**" or "__"). The inner text must not contain
// the delimiter itself.
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

// matchSingleDelim returns the inner text if line is entirely wrapped in a
// single-character delimiter (e.g. '*' or '_') that is not doubled. The inner
// text must not contain the delimiter character.
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

// emphasisContent extracts the inner text if line is entirely a single bold or
// italic span. Returns ("", false) when the line is not a bare emphasis span.
//
// Recognized forms (CommonMark):
//
//	**text**   __text__   (strong)
//	*text*     _text_     (emphasis)
//
// The span must cover the entire trimmed line — no surrounding text allowed.
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

// punctuationChars is the set of sentence-ending punctuation characters that
// indicate the emphasis is prose rather than a heading substitute.
// Includes ASCII and full-width equivalents used in CJK writing systems.
// Mirrors the default punctuation list used by markdownlint MD036.
var punctuationChars = map[rune]struct{}{
	'.': {}, ',': {}, ';': {}, ':': {}, '!': {}, '?': {}, // ASCII
	'。': {}, '、': {}, '；': {}, '：': {}, '！': {}, '？': {}, // Full-width / CJK
}

// endsWithPunctuation reports whether s ends with sentence-ending punctuation.
// When true, the emphasis is likely inline prose, not a heading.
// s is always non-empty when called from CheckNoEmphasisAsHeading.
func endsWithPunctuation(s string) bool {
	runes := []rune(s)
	_, ok := punctuationChars[runes[len(runes)-1]]
	return ok
}

// CheckNoEmphasisAsHeading flags lines where bold or italic text is used as a
// substitute for an ATX heading (MD036). A violation is reported when:
//  1. The trimmed line consists entirely of a single bold/italic span.
//  2. The inner text does not end with sentence-ending punctuation.
//
// Lines inside fenced code blocks and inline code spans are ignored.
func CheckNoEmphasisAsHeading(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		first := firstNonSpaceByte(line)

		if inBlock {
			if first == fenceMarker[0] && IsClosingFence(strings.TrimSpace(line), fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		if first == '`' || first == '~' {
			if marker := openingFenceMarker(strings.TrimSpace(line)); marker != "" {
				inBlock = true
				fenceMarker = marker
			}
			continue
		}

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
