package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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

// isLinkCard reports whether line i is a standalone link-card URL: the
// trimmed line is a single http/https URL with no surrounding prose, preceded
// and followed by a blank line (or the file boundary). Such lines are
// intentionally placed by the author to trigger renderer-level link card
// previews (GitHub, Zenn, etc.) and must not be flagged. The trimmed argument
// is derived from the original line so that inline code or comment markers on
// the line disqualify it from being a bare link card.
func isLinkCard(ctx *preprocess.Context, i int, trimmed string) bool {
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return false
	}
	urls := findBareURLs(trimmed)
	if len(urls) != 1 || trimmed != urls[0] {
		return false
	}
	prevBlank := i == 0 || strings.TrimSpace(ctx.Line(i-1)) == ""
	nextBlank := i >= ctx.Len()-1 || strings.TrimSpace(ctx.Line(i+1)) == ""
	return prevBlank && nextBlank
}

// CheckNoBareURLs flags HTTP/HTTPS URLs that appear as bare text rather than
// being wrapped in angle brackets or used inside a Markdown link or image.
// URLs inside fenced code blocks, indented code blocks, HTML blocks, HTML
// comments, inline code spans, and HTML attribute values are ignored. A URL
// that stands alone on its own line surrounded by blank lines is treated as a
// link card and not flagged.
//
// This rule is the reference adoption of the shared preprocess pass (#337
// Phase 2): rather than re-deriving code/comment context, it skips lines the
// scanner already classified as code or HTML and scans the inline-sanitized
// text (code spans and inline comments blanked) for the remaining lines.
func CheckNoBareURLs(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		// Skip every block context the scanner already identified. This closes
		// the indented-code and HTML-block gaps the previous bespoke fence/
		// comment tracking missed.
		if ctx.InFencedCode(i) || ctx.InIndentedCode(i) || ctx.InHTMLBlock(i) || ctx.InHTMLComment(i) {
			continue
		}

		// Sanitized has inline code spans and inline comments blanked, so URLs
		// living inside them are not seen here.
		sanitized := ctx.Sanitized(i)
		if !strings.Contains(sanitized, "http") {
			continue
		}

		// The link-card test uses the original text: a line carrying anything
		// besides the URL (e.g. an inline code span) is not a bare link card.
		if isLinkCard(ctx, i, strings.TrimSpace(ctx.Line(i))) {
			continue
		}

		for _, url := range findBareURLs(sanitized) {
			errs = append(errs, LintError{
				File:    filename,
				Line:    offset + i + 1,
				Message: fmt.Sprintf("no-bare-urls: bare URL found, use angle brackets or a Markdown link: %s", url),
			})
		}
	}

	return errs
}
