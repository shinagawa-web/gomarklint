package rule

import (
	"fmt"
	"strings"
)

// countBacktickRun returns the number of consecutive backticks starting at
// position start in s.
func countBacktickRun(s string, start int) int {
	n := 0
	for start+n < len(s) && s[start+n] == '`' {
		n++
	}
	return n
}

// stripInlineCode replaces content inside backtick spans (including the
// delimiters) with spaces so that URLs within inline code are not scanned.
// Handles both single-backtick (` “ `) and multi-backtick (` “ `) spans per
// CommonMark: a code span opens with a run of N backticks and closes with the
// next run of exactly N backticks.
func stripInlineCode(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if s[i] != '`' {
			b.WriteByte(s[i])
			i++
			continue
		}

		delimLen := countBacktickRun(s, i)
		closing := -1
		j := i + delimLen
		for j < len(s) {
			if s[j] != '`' {
				j++
				continue
			}
			runLen := countBacktickRun(s, j)
			if runLen == delimLen {
				closing = j
				break
			}
			j += runLen
		}

		if closing == -1 {
			// No matching closing run — emit backticks as-is.
			for k := 0; k < delimLen; k++ {
				b.WriteByte('`')
			}
			i += delimLen
			continue
		}

		// Replace the entire span (delimiters + content) with spaces.
		spanLen := (closing + delimLen) - i
		for k := 0; k < spanLen; k++ {
			b.WriteByte(' ')
		}
		i = closing + delimLen
	}

	return b.String()
}

// CheckNoBareURLs flags HTTP/HTTPS URLs that appear as bare text rather than
// being wrapped in angle brackets or used inside a Markdown link or image.
// URLs inside fenced code blocks and inline code spans are ignored.
func CheckNoBareURLs(filename string, lines []string, offset int) []LintError {
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

		// Fast path: skip lines without any URL scheme.
		if !strings.Contains(line, "http") {
			continue
		}

		scanned := line
		if strings.ContainsRune(line, '`') {
			scanned = stripInlineCode(line)
		}

		// Scan for bare URLs without regex to avoid allocations.
		pos := 0
		for pos < len(scanned) {
			idx := strings.Index(scanned[pos:], "http")
			if idx == -1 {
				break
			}
			start := pos + idx

			// Determine scheme length ("https://" or "http://").
			rest := scanned[start:]
			var schemeLen int
			if strings.HasPrefix(rest, "https://") {
				schemeLen = 8
			} else if strings.HasPrefix(rest, "http://") {
				schemeLen = 7
			} else {
				pos = start + 4
				continue
			}

			// Scan URL body: collect chars matching [^\s<>()\[\]].
			end := start + schemeLen
			for end < len(scanned) {
				c := scanned[end]
				if c <= ' ' || c == '<' || c == '>' || c == '(' || c == ')' || c == '[' || c == ']' {
					break
				}
				end++
			}
			if end == start+schemeLen {
				// No chars after scheme — not a real URL.
				pos = end
				continue
			}

			// Check context: angle-bracket or markdown link destination.
			if start > 0 && scanned[start-1] == '<' {
				pos = end
				continue
			}
			if start > 1 && scanned[start-1] == '(' && scanned[start-2] == ']' {
				pos = end
				continue
			}

			url := strings.TrimRight(scanned[start:end], ".,;:!?)")
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("no-bare-urls: bare URL found, use angle brackets or a Markdown link: %s", url),
			})
			pos = end
		}
	}

	return errs
}
