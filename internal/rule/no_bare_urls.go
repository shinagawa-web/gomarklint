package rule

import (
	"fmt"
	"regexp"
	"strings"
)

var bareURLRegex = regexp.MustCompile(`https?://[^\s<>()\[\]]+`)

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
// Handles both single-backtick (` `` `) and multi-backtick (` `` `) spans per
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

		scanned := stripInlineCode(line)
		matches := bareURLRegex.FindAllStringIndex(scanned, -1)
		for _, m := range matches {
			start := m[0]
			if start > 0 && scanned[start-1] == '<' {
				continue // angle-bracket URL: <https://...>
			}
			if start > 1 && scanned[start-1] == '(' && scanned[start-2] == ']' {
				continue // Markdown link/image destination: ](url)
			}
			url := strings.TrimRight(scanned[m[0]:m[1]], ".,;:!?)")
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("no-bare-urls: bare URL found, use angle brackets or a Markdown link: %s", url),
			})
		}
	}

	return errs
}
