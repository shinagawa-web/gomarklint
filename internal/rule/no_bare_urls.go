package rule

import (
	"fmt"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func isURLBodyChar(c byte) bool {
	return c > ' ' && c != '<' && c != '>' && c != '(' && c != ')' && c != '[' && c != ']' && c != '"' && c != '\''
}

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

func scanURLEnd(line string, bodyStart int) int {
	end := bodyStart
	for end < len(line) && isURLBodyChar(line[end]) {
		end++
	}
	return end
}

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

// isLinkCard: a standalone URL surrounded by blank lines is a renderer link-card preview (GitHub, Zenn, etc.) — not a violation.
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

func CheckNoBareURLs(filename string, ctx *preprocess.Context, offset int) []LintError {
	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if ctx.InFencedCode(i) || ctx.InIndentedCode(i) || ctx.InHTMLBlock(i) || ctx.InHTMLComment(i) {
			continue
		}

		sanitized := ctx.Sanitized(i)
		if !strings.Contains(sanitized, "http") {
			continue
		}

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
