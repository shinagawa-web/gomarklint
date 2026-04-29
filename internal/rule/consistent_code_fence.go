package rule

import (
	"fmt"
	"strings"
)

// CheckConsistentCodeFence flags fenced code blocks that use a different fence
// character than expected. style must be "consistent", "backtick", or "tilde";
// any other value falls back to "consistent".
//
// In "consistent" mode the first fence character found in the document sets the
// expected style; every subsequent opener that differs is flagged.
// In "backtick"/"tilde" mode every opener using the wrong character is flagged.
func CheckConsistentCodeFence(filename string, lines []string, offset int, style string) []LintError {
	switch style {
	case "consistent", "backtick", "tilde":
	default:
		style = "consistent"
	}

	var errs []LintError
	inBlock := false
	fenceMarker := ""
	inHTMLComment := false
	var expectedCh byte // 0 until first fence seen (consistent mode)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inBlock {
			if skip, stillInComment := stepHTMLComment(trimmed, inHTMLComment); skip {
				inHTMLComment = stillInComment
				continue
			}
		}

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		marker := openingFenceMarker(trimmed)
		if marker == "" {
			continue
		}

		ch := marker[0]
		inBlock = true
		fenceMarker = marker

		switch style {
		case "consistent":
			if expectedCh == 0 {
				expectedCh = ch
			} else if ch != expectedCh {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: fmt.Sprintf("consistent-code-fence: expected '%s' fence, got '%s' fence", fenceCharStr(expectedCh), fenceCharStr(ch)),
				})
			}
		case "backtick":
			if ch != '`' {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "consistent-code-fence: expected '```' fence, got '~~~' fence",
				})
			}
		case "tilde":
			if ch != '~' {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "consistent-code-fence: expected '~~~' fence, got '```' fence",
				})
			}
		}
	}

	return errs
}

func fenceCharStr(ch byte) string {
	if ch == '`' {
		return "```"
	}
	return "~~~"
}
