package rule

import (
	"strings"
)

// CheckConsistentCodeFence flags fenced code blocks that use a different fence
// character than expected. style must be "consistent", "backtick", or "tilde".
//
// In "consistent" mode the first fence character found in the document sets the
// expected style; every subsequent opener that differs is flagged.
// In "backtick"/"tilde" mode every opener using the wrong character is flagged.
func CheckConsistentCodeFence(filename string, lines []string, offset int, style string) []LintError {
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
			if err := checkFenceStyle(filename, offset+i+1, ch, style, &expectedCh); err != nil {
				errs = append(errs, *err)
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

// checkFenceStyle validates ch against the configured style and updates
// expectedCh in consistent mode. Returns a LintError if the fence character
// does not match, nil otherwise.
func checkFenceStyle(filename string, line int, ch byte, style string, expectedCh *byte) *LintError {
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
				Message: "consistent-code-fence: expected " + fenceCharName(*expectedCh) + " fence, got " + fenceCharName(ch) + " fence",
			}
		}
	case "backtick":
		if ch != '`' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-code-fence: expected backtick fence, got tilde fence",
			}
		}
	case "tilde":
		if ch != '~' {
			return &LintError{
				File:    filename,
				Line:    line,
				Message: "consistent-code-fence: expected tilde fence, got backtick fence",
			}
		}
	}
	return nil
}

func fenceCharName(ch byte) string {
	if ch == '`' {
		return "backtick"
	}
	return "tilde"
}
