package rule

import "strings"

// CheckNoTrailingSpaces flags lines that end with one or more space or tab
// characters. Lines inside fenced code blocks are ignored.
func CheckNoTrailingSpaces(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		if inBlock {
			if IsClosingFence(strings.TrimSpace(line), fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}

		// Only compute TrimSpace for potential fence opener lines (starts with ` or ~).
		if len(line) >= 3 && (line[0] == '`' || line[0] == '~') {
			if marker := openingFenceMarker(strings.TrimSpace(line)); marker != "" {
				inBlock = true
				fenceMarker = marker
				continue
			}
		}

		if len(line) > 0 {
			last := line[len(line)-1]
			if last == ' ' || last == '\t' {
				errs = append(errs, LintError{
					File:    filename,
					Line:    offset + i + 1,
					Message: "no-trailing-spaces: trailing whitespace found",
				})
			}
		}
	}

	return errs
}
