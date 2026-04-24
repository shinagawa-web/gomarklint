package rule

import (
	"fmt"
	"strings"
)

// CheckNoHardTabs flags hard tab characters (\t) outside fenced code blocks
// and inline code spans. Each tab is reported as a separate violation.
func CheckNoHardTabs(filename string, lines []string, offset int) []LintError {
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

		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		if !strings.ContainsRune(line, '\t') {
			continue
		}

		scanned := line
		if strings.ContainsRune(line, '`') {
			scanned = stripInlineCode(line)
		}

		col := 0
		for _, ch := range scanned {
			col++
			if ch == '\t' {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: fmt.Sprintf("no-hard-tabs: hard tab character found at column %d", col),
				})
			}
		}
	}

	return errs
}
