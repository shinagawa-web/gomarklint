package rule

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

func CheckExternalLinks(path string, content string, skipPatterns []*regexp.Regexp, timeoutSeconds int, urlCache *sync.Map) []LintError {
	codeBlockRanges, _ := GetCodeBlockLineRanges(content)
	links := parser.ExtractExternalLinksWithLineNumbers(content)

	urlToLines := make(map[string][]int)
	for _, link := range links {
		if isInCodeBlock(link.Line, codeBlockRanges) || shouldSkipLink(link.URL, skipPatterns) {
			continue
		}
		urlToLines[link.URL] = append(urlToLines[link.URL], link.Line)
	}

	var errs []LintError
	var mu sync.Mutex
	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	for u, lines := range urlToLines {
		wg.Add(1)
		go func(url string, lns []int) {
			defer wg.Done()

			var status int
			var err error

			if cachedStatus, ok := urlCache.Load(url); ok {
				status = cachedStatus.(int)
			} else {
				status, err = checkURL(client, url)
				urlCache.Store(url, status)
			}

			if err != nil || status >= 400 {
				mu.Lock()
				for _, line := range lns {
					errs = append(errs, LintError{
						File:    path,
						Line:    line,
						Message: formatLinkError(url),
					})
				}
				mu.Unlock()
			}
		}(u, lines)
	}

	wg.Wait()
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
