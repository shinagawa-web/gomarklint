package rule

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

func CheckExternalLinks(path string, content string, skipPatterns []*regexp.Regexp) []LintError {
	codeBlockRanges, _ := GetCodeBlockLineRanges(content)
	links := parser.ExtractExternalLinksWithLineNumbers(content)
	var errs []LintError

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, link := range links {
		if isInCodeBlock(link.Line, codeBlockRanges) {
			continue
		}
		if shouldSkipLink(link.URL, skipPatterns) {
			continue
		}

		status, err := checkURL(client, link.URL)
		if err != nil || status >= 400 {
			errs = append(errs, LintError{
				File:    path,
				Line:    link.Line,
				Message: formatLinkError(link.URL),
			})
		}
	}
	return errs
}

func checkURL(client *http.Client, url string) (int, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err == nil {
		defer func() {
			_ = resp.Body.Close()
		}()
		return resp.StatusCode, nil
	}

	// fallback to GET
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err = client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode, nil
}

func formatLinkError(url string) string {
	return fmt.Sprintf("Link unreachable: %s", url)
}

func shouldSkipLink(url string, skipPatterns []*regexp.Regexp) bool {
	for _, re := range skipPatterns {
		if re.MatchString(url) {
			return true
		}
	}
	return false
}

func isInCodeBlock(line int, ranges [][2]int) bool {
	for _, r := range ranges {
		if line >= r[0] && line <= r[1] {
			return true
		}
	}
	return false
}
