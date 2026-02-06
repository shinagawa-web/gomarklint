package rule

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"
)

var (
	// Link pattern matchers
	inlineLinkPattern = regexp.MustCompile(`\[[^\]]*\]\((https?://[^\s)]+)\)`)
	imageLinkPattern  = regexp.MustCompile(`!\[[^\]]*\]\((https?://[^\s)]+)\)`)
	bareURLPattern    = regexp.MustCompile(`(?m)^.*?(https?://[^\s<>()]+).*?$`)
)

type cacheResult struct {
	status int
	err    error
}

// ExtractedLink represents an external link found in markdown content
type ExtractedLink struct {
	URL  string
	Line int
}

const (
	// DefaultRetryDelayMs is the default delay in milliseconds before retrying a failed HTTP request
	DefaultRetryDelayMs = 1000
)

// ExtractExternalLinksWithLineNumbers extracts external links from the given lines.
// The offset parameter is added to line numbers to account for stripped frontmatter.
func ExtractExternalLinksWithLineNumbers(lines []string, offset int) []ExtractedLink {
	patterns := []*regexp.Regexp{
		inlineLinkPattern,
		imageLinkPattern,
		bareURLPattern,
	}

	var results []ExtractedLink

	for i, line := range lines {
		seenInLine := make(map[string]bool) // Track URLs found in this line
		for _, re := range patterns {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					url := match[1]
					// Only add if not already seen in this line
					if !seenInLine[url] {
						results = append(results, ExtractedLink{
							URL:  url,
							Line: i + 1 + offset, // 1-based line number + offset for frontmatter
						})
						seenInLine[url] = true
					}
				}
			}
		}
	}
	return results
}

// CheckExternalLinks checks external links in the given lines.
// The offset parameter is used to calculate correct line numbers accounting for stripped frontmatter.
func CheckExternalLinks(path string, lines []string, offset int, skipPatterns []*regexp.Regexp, timeoutSeconds int, retryDelayMs int, urlCache *sync.Map) []LintError {
	codeBlockRanges, _ := GetCodeBlockLineRanges(lines)
	links := ExtractExternalLinksWithLineNumbers(lines, offset)

	urlToLines := make(map[string][]int)
	for _, link := range links {
		// Adjust line number for code block check (relative to lines array)
		relativeLineNum := link.Line - offset
		if isInCodeBlock(relativeLineNum, codeBlockRanges) || shouldSkipLink(link.URL, skipPatterns) {
			continue
		}
		urlToLines[link.URL] = append(urlToLines[link.URL], link.Line)
	}

	var errs []LintError
	var mu sync.Mutex
	var wg sync.WaitGroup

	// maxConcurrency limits the number of concurrent HTTP requests
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	for u, lines := range urlToLines {
		wg.Add(1)
		sem <- struct{}{}
		go func(url string, lns []int) {
			defer wg.Done()
			defer func() { <-sem }()
			var status int
			var err error

			if cached, ok := urlCache.Load(url); ok {
				if result, ok := cached.(cacheResult); ok {
					status = result.status
					err = result.err
				} else {
					// Cache contained unexpected type, re-check the URL
					status, err = checkURL(client, url, retryDelayMs)
					urlCache.Store(url, cacheResult{status: status, err: err})
				}
			} else {
				status, err = checkURL(client, url, retryDelayMs)
				urlCache.Store(url, cacheResult{status: status, err: err})
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

// checkURL performs the URL check with retry logic.
func checkURL(client *http.Client, url string, retryDelayMs int) (int, error) {
	// maxRetries is the maximum number of retry attempts for failed requests
	const maxRetries = 2
	retryDelay := time.Duration(retryDelayMs) * time.Millisecond

	var status int
	var err error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// Wait a bit before retrying (simple backoff)
			time.Sleep(retryDelay * time.Duration(i))
		}

		status, err = performCheck(client, url)

		// Success: 2xx or 3xx
		if err == nil && status < 400 {
			return status, nil
		}

		// Permanent failures: Don't bother retrying if it's 404 (Not Found) or 401 (Unauthorized)
		if err == nil && (status == http.StatusNotFound || status == http.StatusUnauthorized) {
			return status, nil
		}

		// If it's the last attempt, don't log "retrying"
		if i == maxRetries {
			break
		}

		// Optional: You could log that you're retrying here
	}

	return status, err
}

// performCheck contains the core HEAD -> GET fallback logic.
func performCheck(client *http.Client, url string) (int, error) {
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
