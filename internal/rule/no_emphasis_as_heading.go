package rule

import (
	"fmt"
	"strings"
)

// emphasisContent extracts the inner text if line is entirely a single bold or
// italic span. Returns ("", false) when the line is not a bare emphasis span.
//
// Recognised forms (CommonMark):
//
//	**text**   __text__   (strong)
//	*text*     _text_     (emphasis)
//
// The span must cover the entire trimmed line — no surrounding text allowed.
func emphasisContent(line string) (string, bool) {
	// Strong: **...**
	if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") && len(line) > 4 {
		inner := line[2 : len(line)-2]
		if !strings.Contains(inner, "**") {
			return inner, true
		}
	}
	// Strong: __...__
	if strings.HasPrefix(line, "__") && strings.HasSuffix(line, "__") && len(line) > 4 {
		inner := line[2 : len(line)-2]
		if !strings.Contains(inner, "__") {
			return inner, true
		}
	}
	// Emphasis: *...*  (not **)
	if len(line) > 2 && line[0] == '*' && line[len(line)-1] == '*' &&
		!strings.HasPrefix(line, "**") {
		inner := line[1 : len(line)-1]
		if !strings.Contains(inner, "*") {
			return inner, true
		}
	}
	// Emphasis: _..._  (not __)
	if len(line) > 2 && line[0] == '_' && line[len(line)-1] == '_' &&
		!strings.HasPrefix(line, "__") {
		inner := line[1 : len(line)-1]
		if !strings.Contains(inner, "_") {
			return inner, true
		}
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
func endsWithPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
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
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		// Strip inline code before checking so that e.g. `**bold**` is ignored.
		scanned := trimmed
		if strings.ContainsRune(trimmed, '`') {
			scanned = strings.TrimSpace(stripInlineCode(trimmed))
		}

		inner, ok := emphasisContent(scanned)
		if !ok {
			continue
		}
		if endsWithPunctuation(inner) {
			continue
		}

		errs = append(errs, LintError{
			File:    filename,
			Line:    offset + i + 1,
			Message: fmt.Sprintf("no-emphasis-as-heading: emphasis used as heading, use ATX heading instead: %s", scanned),
		})
	}

	return errs
}
