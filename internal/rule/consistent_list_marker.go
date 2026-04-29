package rule

import (
	"strings"
)

// CheckConsistentListMarker flags unordered list items that use a different
// marker than expected. style must be "consistent", "dash", "asterisk", or
// "plus"; any other value falls back to "consistent".
//
// In "consistent" mode the first marker found in the document sets the
// expected style; every subsequent item using a different marker is flagged.
// In "dash"/"asterisk"/"plus" mode every item using the wrong marker is
// flagged. Content inside fenced code blocks is ignored.
func CheckConsistentListMarker(filename string, lines []string, offset int, style string) []LintError {
	switch style {
	case "consistent", "dash", "asterisk", "plus":
	default:
		style = "consistent"
	}

	var errs []LintError
	inBlock := false
	fenceMarker := ""
	var expectedCh byte // 0 until first list item seen (consistent mode)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		first := firstNonSpaceByte(line)

		if first == '`' || first == '~' {
			if marker := openingFenceMarker(trimmed); marker != "" {
				inBlock = true
				fenceMarker = marker
				continue
			}
		}

		if first != '-' && first != '*' && first != '+' {
			continue
		}

		ch, ok := listItemMarker(line)
		if !ok {
			continue
		}

		if err := checkListMarkerStyle(filename, offset+i+1, ch, style, &expectedCh); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

// listItemMarker returns the marker byte and true if line is an unordered list
// item (optional leading spaces, one of - * +, then a space and non-space content).
// Returns 0, false otherwise.
func listItemMarker(line string) (byte, bool) {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	ch := line[i]
	i++
	if i >= len(line) || line[i] != ' ' {
		return 0, false
	}
	i++
	if i >= len(line) || line[i] == ' ' || line[i] == '\t' || line[i] == '\n' {
		return 0, false
	}
	return ch, true
}

// checkListMarkerStyle validates ch against the configured style and updates
// expectedCh in consistent mode. Returns a LintError if the marker does not
// match, nil otherwise.
func checkListMarkerStyle(filename string, line int, ch byte, style string, expectedCh *byte) *LintError {
	switch style {
	case "consistent":
		if *expectedCh == 0 {
			*expectedCh = ch
			return nil
		}
		if ch != *expectedCh {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected " + listMarkerName(*expectedCh) + " marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "dash":
		if ch != '-' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected dash marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "asterisk":
		if ch != '*' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected asterisk marker, got " + listMarkerName(ch) + " marker",
			}
		}
	case "plus":
		if ch != '+' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-list-marker: expected plus marker, got " + listMarkerName(ch) + " marker",
			}
		}
	}
	return nil
}

func listMarkerName(ch byte) string {
	switch ch {
	case '-':
		return "dash"
	case '*':
		return "asterisk"
	default:
		return "plus"
	}
}
