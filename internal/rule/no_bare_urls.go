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

// isURLBodyChar returns true for characters allowed in a URL body
// (everything except whitespace and <>()[].)
func isURLBodyChar(c byte) bool {
	return c > ' ' && c != '<' && c != '>' && c != '(' && c != ')' && c != '[' && c != ']'
}

// isWrappedURL returns true if the URL starting at start in line is wrapped
// in angle brackets, is a Markdown link/image destination, or appears inside
// an HTML attribute value (quoted with " or ').
func isWrappedURL(line string, start int) bool {
	if start > 0 && line[start-1] == '<' {
		return true
	}
	if start > 1 && line[start-1] == '(' && line[start-2] == ']' {
		return true
	}
	return start > 0 && (line[start-1] == '"' || line[start-1] == '\'')
}

// stripHTMLComments replaces content inside <!-- ... --> spans (including the
// delimiters) with spaces so that URLs within HTML comments are not scanned.
// Only handles comments that open and close on the same line.
func stripHTMLComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if i+4 <= len(s) && s[i:i+4] == "<!--" {
			end := strings.Index(s[i+4:], "-->")
			if end == -1 {
				// Unclosed comment — blank the rest of the line.
				for k := i; k < len(s); k++ {
					b.WriteByte(' ')
				}
				return b.String()
			}
			spanLen := 4 + end + 3
			for k := 0; k < spanLen; k++ {
				b.WriteByte(' ')
			}
			i += spanLen
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

// scanURLEnd returns the end index of the URL body starting at bodyStart.
func scanURLEnd(line string, bodyStart int) int {
	end := bodyStart
	for end < len(line) && isURLBodyChar(line[end]) {
		end++
	}
	return end
}

// findBareURLs scans line for bare HTTP/HTTPS URLs and returns their text.
// URLs wrapped in angle brackets or used as Markdown link destinations are
// skipped.
func findBareURLs(line string) []string {
	var urls []string
	pos := 0
	for pos < len(line) {
		idx := strings.Index(line[pos:], "http")
		if idx == -1 {
			break
		}
		start := pos + idx

		// Determine scheme length ("https://" or "http://").
		rest := line[start:]
		var schemeLen int
		if strings.HasPrefix(rest, "https://") {
			schemeLen = 8
		} else if strings.HasPrefix(rest, "http://") {
			schemeLen = 7
		} else {
			pos = start + 4
			continue
		}

		end := scanURLEnd(line, start+schemeLen)
		if end == start+schemeLen || isWrappedURL(line, start) {
			pos = end
			continue
		}

		urls = append(urls, strings.TrimRight(line[start:end], ".,;:!?)"))
		pos = end
	}
	return urls
}

// CheckNoBareURLs flags HTTP/HTTPS URLs that appear as bare text rather than
// being wrapped in angle brackets or used inside a Markdown link or image.
// URLs inside fenced code blocks and inline code spans are ignored.
func CheckNoBareURLs(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	inComment := false

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

		// Multi-line HTML comment tracking.
		if inComment {
			if idx := strings.Index(line, "-->"); idx != -1 {
				inComment = false
				line = strings.Repeat(" ", idx+3) + line[idx+3:]
			} else {
				continue
			}
		}

		if !strings.Contains(line, "http") {
			// Still check for comment open even on non-http lines.
			if strings.Contains(line, "<!--") && !strings.Contains(line, "-->") {
				inComment = true
			}
			continue
		}

		scanned := line
		if strings.ContainsRune(scanned, '`') {
			scanned = stripInlineCode(scanned)
		}
		if strings.Contains(scanned, "<!--") {
			if !strings.Contains(scanned[strings.Index(scanned, "<!--")+4:], "-->") {
				inComment = true
			}
			scanned = stripHTMLComments(scanned)
		}

		for _, url := range findBareURLs(scanned) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("no-bare-urls: bare URL found, use angle brackets or a Markdown link: %s", url),
			})
		}
	}

	return errs
}
