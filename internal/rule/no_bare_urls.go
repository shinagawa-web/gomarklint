package rule

import (
	"fmt"
	"regexp"
	"strings"
)

var bareURLRegex = regexp.MustCompile(`https?://[^\s<>()\[\]]+`)

// stripInlineCode replaces content inside backtick spans with spaces so that
// URLs within inline code are not scanned.
func stripInlineCode(s string) string {
	var b strings.Builder
	inCode := false
	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			inCode = !inCode
			b.WriteByte(' ')
		} else if inCode {
			b.WriteByte(' ')
		} else {
			b.WriteByte(s[i])
		}
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
			if start > 0 {
				prev := scanned[start-1]
				if prev == '(' || prev == '<' {
					continue
				}
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
