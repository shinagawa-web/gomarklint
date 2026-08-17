package rule

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

var (
	inlineLinkPattern = regexp.MustCompile(`\[[^\]]*\]\((https?://(?:[^\s()]+|\([^\s()]*\))+)\)`)
	imageLinkPattern  = regexp.MustCompile(`!\[[^\]]*\]\((https?://(?:[^\s()]+|\([^\s()]*\))+)\)`)
	bareURLPattern    = regexp.MustCompile(`(?m)^.*?(https?://(?:[^\s<>"'()]+|\([^\s<>"'()]*\))+).*?$`)
)

type cacheResult struct {
	status int
	err    error
}

type ExtractedLink struct {
	URL  string
	Line int
}

const (
	DefaultRetryDelayMs        = 1000
	DefaultMaxConcurrency      = 10
	MaxConcurrencyLimit        = 15
	DefaultMaxRetries          = 2
	MaxRetriesLimit            = 4
	DefaultPerHostConcurrency  = 2
	MaxPerHostConcurrencyLimit = 15
	DefaultPerHostIntervalMs   = 3000
	MinPerHostIntervalMs       = 1000
	MaxPerHostIntervalMsLimit  = 60000

	userAgent = "gomarklint/v3 (+https://github.com/shinagawa-web/gomarklint)"
)

// 429 (Too Many Requests) is rate limiting, not a broken link.
var defaultAllowedStatuses = []int{http.StatusTooManyRequests}

func isAllowedStatus(status int, extra []int) bool {
	for _, s := range defaultAllowedStatuses {
		if status == s {
			return true
		}
	}
	for _, s := range extra {
		if status == s {
			return true
		}
	}
	return false
}

type hostLimiter struct {
	sem       chan struct{} // nil when perHostConcurrency == 0
	interval  time.Duration
	nextAvail time.Time
	mu        sync.Mutex
}

func (h *hostLimiter) acquire() {
	if h.sem != nil {
		h.sem <- struct{}{}
	}
	if h.interval <= 0 {
		return
	}
	h.mu.Lock()
	now := time.Now()
	var wait time.Duration
	if h.nextAvail.After(now) {
		wait = h.nextAvail.Sub(now)
		h.nextAvail = h.nextAvail.Add(h.interval)
	} else {
		h.nextAvail = now.Add(h.interval)
	}
	h.mu.Unlock()
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (h *hostLimiter) release() {
	if h.sem != nil {
		<-h.sem
	}
}

type hostLimiterRegistry struct {
	mu                 sync.Mutex
	limiters           map[string]*hostLimiter
	perHostConcurrency int
	perHostIntervalMs  int
}

func newHostLimiterRegistry(perHostConcurrency, perHostIntervalMs int) *hostLimiterRegistry {
	return &hostLimiterRegistry{
		limiters:           make(map[string]*hostLimiter),
		perHostConcurrency: perHostConcurrency,
		perHostIntervalMs:  perHostIntervalMs,
	}
}

func (r *hostLimiterRegistry) get(host string) *hostLimiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if lim, ok := r.limiters[host]; ok {
		return lim
	}
	lim := &hostLimiter{interval: time.Duration(r.perHostIntervalMs) * time.Millisecond}
	if r.perHostConcurrency > 0 {
		lim.sem = make(chan struct{}, r.perHostConcurrency)
	}
	r.limiters[host] = lim
	return lim
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}

func ExtractExternalLinksWithLineNumbers(ctx *preprocess.Context, offset int) []ExtractedLink {
	patterns := []*regexp.Regexp{
		inlineLinkPattern,
		imageLinkPattern,
		bareURLPattern,
	}

	var results []ExtractedLink

	for i := 0; i < ctx.Len(); i++ {
		// Skip code/HTML block contexts entirely, and scan the inline-sanitized
		// text so URLs inside inline code spans and inline comments are not
		// extracted (and therefore not fetched).
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Sanitized(i)
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
							Line: i + 1 + offset,
						})
						seenInLine[url] = true
					}
				}
			}
		}
	}
	return results
}

func CheckExternalLinks(path string, ctx *preprocess.Context, offset int, skipPatterns []*regexp.Regexp, timeoutSeconds int, retryDelayMs int, maxConcurrency int, maxRetries int, allowedStatuses []int, urlCache *sync.Map, perHostConcurrency int, perHostIntervalMs int) ([]LintError, int) {
	links := ExtractExternalLinksWithLineNumbers(ctx, offset)

	urlToLines := make(map[string][]int)
	for _, link := range links {
		if shouldSkipLink(link.URL, skipPatterns) {
			continue
		}
		urlToLines[link.URL] = append(urlToLines[link.URL], link.Line)
	}

	var errs []LintError
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, maxConcurrency)

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	hostReg := newHostLimiterRegistry(perHostConcurrency, perHostIntervalMs)

	for u, lines := range urlToLines {
		wg.Add(1)
		sem <- struct{}{}
		go func(url string, lns []int) {
			defer wg.Done()
			defer func() { <-sem }()

			var status int
			var err error

			needHTTP := true
			if cached, ok := urlCache.Load(url); ok {
				if result, ok := cached.(cacheResult); ok {
					status = result.status
					err = result.err
					needHTTP = false
				}
			}

			if needHTTP {
				lim := hostReg.get(extractHost(url))
				lim.acquire()
				defer lim.release()
				status, err = checkURL(client, url, retryDelayMs, maxRetries, allowedStatuses)
				urlCache.Store(url, cacheResult{status: status, err: err})
			}

			if err != nil || (status >= 400 && !isAllowedStatus(status, allowedStatuses)) {
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
	return errs, len(urlToLines)
}

func checkURL(client *http.Client, url string, retryDelayMs int, maxRetries int, allowedStatuses []int) (int, error) {
	retryDelay := time.Duration(retryDelayMs) * time.Millisecond

	var status int
	var err error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay * time.Duration(1<<uint(i-1)))
		}

		status, err = performCheck(client, url)

		if err == nil && status < 400 {
			return status, nil
		}
		if err == nil && isAllowedStatus(status, allowedStatuses) {
			return status, nil
		}
		if err == nil && (status == http.StatusNotFound || status == http.StatusUnauthorized) {
			return status, nil
		}
		if i == maxRetries {
			break
		}
	}

	return status, err
}

func performCheck(client *http.Client, url string) (int, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		// Some servers reject HEAD with 405 (Method Not Allowed) or 403 (Forbidden)
		// but serve GET normally. Fall back to GET in those cases.
		if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusForbidden {
			return resp.StatusCode, nil
		}
	}

	// fallback to GET: covers both network errors and HEAD 405/403
	req.Method = "GET"

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
