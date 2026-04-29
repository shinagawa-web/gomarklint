package rule

import (
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

		// Inside a code block: only look for the closing fence.
		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		// Inside an HTML comment block: skip until "-->" is found.
		if inHTMLComment {
			if strings.Contains(trimmed, "-->") {
				inHTMLComment = false
			}
			continue
		}

		// Check for an opening fence before inspecting the line for "<!--" so
		// that fence openers whose info string contains "<!--" (e.g.
		// "```go <!-- note -->") are treated as fences, not comment starts.
		if marker := openingFenceMarker(trimmed); marker != "" {
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
						Message: "consistent-code-fence: expected " + fenceCharName(expectedCh) + " fence, got " + fenceCharName(ch) + " fence",
					})
				}
			case "backtick":
				if ch != '`' {
					errs = append(errs, LintError{
						File:    filename,
						Line:    offset + i + 1,
						Message: "consistent-code-fence: expected backtick fence, got tilde fence",
					})
				}
			case "tilde":
				if ch != '~' {
					errs = append(errs, LintError{
						File:    filename,
						Line:    offset + i + 1,
						Message: "consistent-code-fence: expected tilde fence, got backtick fence",
					})
				}
			}
			continue
		}

		// Track HTML comment blocks so fences inside them are ignored.
		if strings.Contains(trimmed, "<!--") && !strings.Contains(trimmed, "-->") {
			inHTMLComment = true
		}
	}

	return errs
}

func fenceCharName(ch byte) string {
	if ch == '`' {
		return "backtick"
	}
	return "tilde"
}
