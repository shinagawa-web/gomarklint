package rule

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

func CheckExternalLinks(path string, content string, skipPatterns []*regexp.Regexp) []LintError {
	codeBlockRanges, _ := GetCodeBlockLineRanges(content)
	links := parser.ExtractExternalLinksWithLineNumbers(content)
	var errs []LintError
	var mu sync.Mutex
	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	for _, link := range links {
		if isInCodeBlock(link.Line, codeBlockRanges) {
			continue
		}
		if shouldSkipLink(link.URL, skipPatterns) {
			continue
		}

		wg.Add(1)
		go func(l parser.ExtractedLink) {
			defer wg.Done()

			status, err := checkURL(client, l.URL)
			if err != nil || status >= 400 {
				mu.Lock()
				errs = append(errs, LintError{
					File:    path,
					Line:    l.Line,
					Message: formatLinkError(l.URL),
				})
				mu.Unlock()
			}
		}(link)
	}

	wg.Wait()

	// Sort errors by line number to ensure consistent output
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Line < errs[j].Line
	})

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
