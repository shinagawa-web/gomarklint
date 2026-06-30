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
	// Link pattern matchers
	// Inline/image patterns allow balanced (...) groups inside the URL per CommonMark spec.
	// [^\s()]+ allows square brackets (needed for IPv6 hosts like https://[::1]/).
	// + quantifier ensures at least one character after the scheme (rejects bare https://).
	inlineLinkPattern = regexp.MustCompile(`\[[^\]]*\]\((https?://(?:[^\s()]+|\([^\s()]*\))+)\)`)
	imageLinkPattern  = regexp.MustCompile(`!\[[^\]]*\]\((https?://(?:[^\s()]+|\([^\s()]*\))+)\)`)
	bareURLPattern    = regexp.MustCompile(`(?m)^.*?(https?://(?:[^\s<>"'()]+|\([^\s<>"'()]*\))+).*?$`)
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
	// DefaultMaxConcurrency is the default maximum number of concurrent HTTP requests
	DefaultMaxConcurrency = 10
	// MaxConcurrencyLimit is the maximum allowed value for maxConcurrency
	MaxConcurrencyLimit = 15
	// DefaultMaxRetries is the default maximum number of retry attempts for failed requests
	DefaultMaxRetries = 2
	// MaxRetriesLimit is the maximum allowed value for maxRetries
	MaxRetriesLimit = 4

	// DefaultPerHostConcurrency is the default per-host concurrency limit
	DefaultPerHostConcurrency = 2
	// MaxPerHostConcurrencyLimit is the maximum allowed value for perHostConcurrency
	MaxPerHostConcurrencyLimit = 15
	// DefaultPerHostIntervalMs is the default minimum interval between requests to the same host
	DefaultPerHostIntervalMs = 3000
	// MinPerHostIntervalMs is the minimum non-zero value for perHostIntervalMs; values between 1 and 999 are rejected
	MinPerHostIntervalMs = 1000
	// MaxPerHostIntervalMsLimit is the maximum allowed value for perHostIntervalMs in milliseconds
	MaxPerHostIntervalMsLimit = 60000

	userAgent = "gomarklint/v3 (+https://github.com/shinagawa-web/gomarklint)"
)

// defaultAllowedStatuses contains status codes never treated as link failures.
// 429 (Too Many Requests) indicates rate limiting, not a broken link.
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

// hostLimiter enforces per-host concurrency and minimum request interval.
type hostLimiter struct {
	sem       chan struct{} // nil when perHostConcurrency == 0
	interval  time.Duration
	nextAvail time.Time
	mu        sync.Mutex
}

// acquire waits for a per-host slot and enforces the minimum interval before returning.
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

// release frees the per-host semaphore slot.
func (h *hostLimiter) release() {
	if h.sem != nil {
		<-h.sem
	}
}

// hostLimiterRegistry maintains one hostLimiter per host.
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

// extractHost returns the host component of rawURL (e.g. "github.com").
func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}

// ExtractExternalLinksWithLineNumbers extracts external links from the given lines.
// The offset parameter is added to line numbers to account for stripped frontmatter.
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
// Returns lint errors and the count of unique URLs checked.
func CheckExternalLinks(path string, ctx *preprocess.Context, offset int, skipPatterns []*regexp.Regexp, timeoutSeconds int, retryDelayMs int, maxConcurrency int, maxRetries int, allowedStatuses []int, urlCache *sync.Map, perHostConcurrency int, perHostIntervalMs int) ([]LintError, int) {
	// Code/HTML context filtering is handled inside the extractor via the shared
	// scanner, so links inside fenced/indented code, HTML blocks, HTML comments,
	// inline code spans, and inline comments are never produced here.
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

// checkURL performs the URL check with retry logic.
func checkURL(client *http.Client, url string, retryDelayMs int, maxRetries int, allowedStatuses []int) (int, error) {
	retryDelay := time.Duration(retryDelayMs) * time.Millisecond

	var status int
	var err error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay * time.Duration(1<<uint(i-1)))
		}

		status, err = performCheck(client, url)

		// Success: 2xx or 3xx
		if err == nil && status < 400 {
			return status, nil
		}

		// Allowed statuses (e.g. 429, user-configured): return immediately without retrying
		if err == nil && isAllowedStatus(status, allowedStatuses) {
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
