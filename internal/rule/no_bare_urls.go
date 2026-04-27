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
// Handles both single-backtick (` " `) and multi-backtick (` " `) spans per
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
// (everything except whitespace, <>, ()[], and quotes).
func isURLBodyChar(c byte) bool {
	return c > ' ' && c != '<' && c != '>' && c != '(' && c != ')' && c != '[' && c != ']' && c != '"' && c != '\''
}

// isWrappedURL returns true if the URL starting at start in line is wrapped
// in angle brackets, is a Markdown link/image destination, or appears inside
// an HTML attribute value (a quote immediately preceded by '=').
func isWrappedURL(line string, start int) bool {
	if start > 0 && line[start-1] == '<' {
		return true
	}
	if start > 1 && line[start-1] == '(' && line[start-2] == ']' {
		return true
	}
	// HTML attribute value: require an '=' before the opening quote,
	// optionally separated by whitespace (e.g. href="..." or attr = "...").
	if start > 0 && (line[start-1] == '"' || line[start-1] == '\'') {
		i := start - 2
		for i >= 0 && (line[i] == ' ' || line[i] == '\t') {
			i--
		}
		return i >= 0 && line[i] == '='
	}
	return false
}

// stripHTMLComments replaces content inside <!-- ... --> spans (including the
// delimiters) with spaces so that URLs within HTML comments are not scanned.
// It handles multiple comment spans on a single line. The second return value
// reports whether the line ended inside an unclosed comment.
func stripHTMLComments(s string) (string, bool) {
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
				return b.String(), true
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
	return b.String(), false
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

// advanceComment advances a line that is currently inside a multi-line HTML
// comment. It returns the (possibly modified) line and the updated inComment
// state. If "-->" is found, inComment becomes false and the remainder of the
// line after "-->" is returned for further processing.
func advanceComment(line string) (string, bool) {
	idx := strings.Index(line, "-->")
	if idx == -1 {
		return line, true
	}
	return strings.Repeat(" ", idx+3) + line[idx+3:], false
}

// prepareScanned strips inline code and HTML comments from line, returning the
// sanitized string and whether the line ended inside an unclosed comment.
func prepareScanned(line string) (string, bool) {
	scanned := line
	if strings.ContainsRune(scanned, '`') {
		scanned = stripInlineCode(scanned)
	}
	if strings.Contains(scanned, "<!--") {
		var endedInComment bool
		scanned, endedInComment = stripHTMLComments(scanned)
		return scanned, endedInComment
	}
	return scanned, false
}

// isLinkCard reports whether line i is a standalone link-card URL: the
// trimmed line is a single http/https URL with no surrounding prose, preceded
// and followed by a blank line (or the file boundary). Such lines are
// intentionally placed by the author to trigger renderer-level link card
// previews (GitHub, Zenn, etc.) and must not be flagged.
func isLinkCard(lines []string, i int, trimmed string) bool {
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return false
	}
	if strings.ContainsAny(trimmed, " \t") {
		return false
	}
	prevBlank := i == 0 || strings.TrimSpace(lines[i-1]) == ""
	nextBlank := i >= len(lines)-1 || strings.TrimSpace(lines[i+1]) == ""
	return prevBlank && nextBlank
}

// CheckNoBareURLs flags HTTP/HTTPS URLs that appear as bare text rather than
// being wrapped in angle brackets or used inside a Markdown link or image.
// URLs inside fenced code blocks, inline code spans, HTML comments, and HTML
// attribute values are ignored. A URL that stands alone on its own line
// surrounded by blank lines is treated as a link card and not flagged.
func CheckNoBareURLs(filename string, lines []string, offset int) []LintError {
	var errs []LintError
	inBlock := false
	fenceMarker := ""
	inComment := false

	for i, line := range lines {
		// Multi-line HTML comment tracking takes priority: fences inside
		// comments are not treated as code block delimiters.
		if inComment {
			var stillInComment bool
			line, stillInComment = advanceComment(line)
			if stillInComment {
				continue
			}
			inComment = false
			// Fall through to process the remainder of the line.
		}

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

		if !strings.Contains(line, "http") {
			if strings.Contains(line, "<!--") {
				_, inComment = stripHTMLComments(line)
			}
			continue
		}

		if isLinkCard(lines, i, trimmed) {
			continue
		}

		scanned, endedInComment := prepareScanned(line)
		if endedInComment {
			inComment = true
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
