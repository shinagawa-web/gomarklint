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

// bodyHasTrailingWhitespace reports whether body contains any line ending with
// a space or tab. It uses strings.IndexByte to locate each '\n' — IndexByte is
// backed by AVX2 assembly on amd64 (1-byte SIMD search), which is far faster
// than strings.Contains for 2-byte patterns on that architecture.
// Total bytes scanned equals len(body): each call advances past the found '\n'.
func bodyHasTrailingWhitespace(body string) bool {
	s := body
	for len(s) > 0 {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			break
		}
		if i > 0 {
			prev := s[i-1]
			if prev == ' ' || prev == '\t' {
				return true
			}
			// CRLF: "text   \r\n" — check byte before the \r
			if prev == '\r' && i >= 2 && (s[i-2] == ' ' || s[i-2] == '\t') {
				return true
			}
		}
		s = s[i+1:]
	}
	// Handle trailing whitespace at EOF (file without a final newline)
	n := len(body)
	return n > 0 && (body[n-1] == ' ' || body[n-1] == '\t')
}

// CheckNoTrailingSpaces flags lines that end with one or more space or tab
// characters. Lines inside fenced code blocks are ignored.
//
// body is the raw document string (before splitting); it is used for a fast-path
// check via bodyHasTrailingWhitespace, which uses IndexByte (AVX2 on amd64) to
// locate newlines efficiently. lines is the same content split on "\n".
func CheckNoTrailingSpaces(filename, body string, lines []string, offset int) []LintError {
	if !bodyHasTrailingWhitespace(body) {
		return nil
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
