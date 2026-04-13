package rule

import "strings"

// stripCR removes a trailing carriage return from a line, if present.
// This normalises CRLF input so that trailing-space detection works correctly
// on files with Windows line endings.
func stripCR(line string) string {
	if len(line) > 0 && line[len(line)-1] == '\r' {
		return line[:len(line)-1]
	}
	return line
}

// CheckNoTrailingSpaces flags lines that end with one or more space or tab
// characters. Lines inside fenced code blocks are ignored.
//
// body is the raw document string (before splitting); it is used for a fast-path
// check using SIMD-optimised string search to skip line-by-line analysis on
// clean documents. lines is the same content split on "\n".
func CheckNoTrailingSpaces(filename, body string, lines []string, offset int) []LintError {
	// Fast path: use strings.Contains which Go's runtime implements with
	// SIMD-optimised search on amd64. Scanning a contiguous string is far
	// cheaper than iterating a []string with pointer indirection per element.
	//
	// " \r" covers CRLF trailing spaces ("text   \r\n" contains " \r").
	// The final length+last-byte check handles a missing trailing newline.
	if !strings.Contains(body, " \n") && !strings.Contains(body, "\t\n") &&
		!strings.Contains(body, " \r") && !strings.Contains(body, "\t\r") {
		n := len(body)
		if n == 0 {
			return nil
		}
		last := body[n-1]
		if last != ' ' && last != '\t' {
			return nil
		}
	}

	var errs []LintError
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		line = stripCR(line)

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
