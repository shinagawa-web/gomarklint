package rule

import (
	"fmt"
	"strings"
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
func CheckMaxLineLength(filename string, lines []string, offset int, lineLength int) []LintError {
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
				continue
			}
		}

		if len(line) <= lineLength {
			continue
		}

		trimmed := strings.TrimSpace(line)
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
