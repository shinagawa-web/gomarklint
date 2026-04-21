package rule

import (
	"fmt"
	"strings"
)

// isBareURLLine reports whether the trimmed line consists solely of a URL
// (http:// or https://), possibly wrapped in angle brackets.
func isBareURLLine(trimmed string) bool {
	s := trimmed
	if strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") {
		s = s[1 : len(s)-1]
	}
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// CheckMaxLineLength flags lines whose byte length exceeds lineLength.
// Lines inside fenced code blocks, ATX heading lines, and lines that consist
// solely of a URL are exempt.
func CheckMaxLineLength(filename string, lines []string, offset int, lineLength int) []LintError {
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

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if isBareURLLine(trimmed) {
			continue
		}

		if len(line) > lineLength {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("max-line-length: line exceeds %d characters (%d)", lineLength, len(line)),
			})
		}
	}

	return errs
}
