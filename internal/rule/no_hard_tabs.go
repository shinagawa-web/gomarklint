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
		first := firstNonSpaceByte(line)

		if inBlock {
			if first == '`' || first == '~' {
				if IsClosingFence(strings.TrimSpace(line), fenceMarker) {
					inBlock = false
					fenceMarker = ""
				}
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

		if strings.IndexByte(line, '\t') < 0 {
			continue
		}

		scanned := line
		if strings.IndexByte(line, '`') >= 0 {
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
