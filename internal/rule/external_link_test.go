package rule_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/rule"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/fail":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
}

// Helper function to convert markdown string to lines and offset
func toLines(markdown string) ([]string, int) {
	return strings.Split(markdown, "\n"), 0
}

func TestCheckExternalLinks_BasicSuccessFailure(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	markdown := fmt.Sprintf(`[ok](%s/ok)
[fail](%s/fail)
`, ts.URL, ts.URL)

	file := "mock.md"
	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(file, lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	if len(results) != 1 {
		t.Fatalf("expected 1 error, got %d", len(results))
	}

	got := results[0]
	if !strings.Contains(got.Message, "/fail") {
		t.Errorf("unexpected error message: %s", got.Message)
	}
	if got.Line != 2 {
		t.Errorf("expected line number 2, got %d", got.Line)
	}
	if got.File != file {
		t.Errorf("expected file '%s', got %s", file, got.File)
	}
}

func TestCheckExternalLinks_SkipPattern(t *testing.T) {
	markdown := `
[skip this](http://localhost/skip)
[check this](https://httpstat.us/404)
`
	skip := []*regexp.Regexp{
		regexp.MustCompile(`localhost`),
	}

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("mock.md", lines, offset, skip, 2, 10, &sync.Map{})

	if len(results) != 1 {
		t.Fatalf("expected 1 error (only non-localhost link should be checked), got %d", len(results))
	}
	if !strings.Contains(results[0].Message, "httpstat.us") {
		t.Errorf("expected error for httpstat.us link, got: %v", results[0])
	}
}

func TestCheckExternalLinks_IgnoreCodeBlocks(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	markdown := fmt.Sprintf("```\n[code link](%s/in-code)\n```\n[real link](%s/fail)\n", ts.URL, ts.URL)

	skip := []*regexp.Regexp{}

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("mock.md", lines, offset, skip, 10, 10, &sync.Map{})
	if len(results) != 1 {
		t.Fatalf("expected 1 error (code block link should be ignored), got %d", len(results))
	}
	if !strings.Contains(results[0].Message, "/fail") {
		t.Errorf("expected error for real link, got: %v", results[0])
	}
	if results[0].Line != 4 {
		t.Errorf("expected line 4 for real link, got: %d", results[0].Line)
	}
}

func TestCheckExternalLinks_ParallelCheck(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	// Create markdown with multiple failing links with different URLs
	// to avoid deduplication in parser
	markdown := fmt.Sprintf(`Line 1 [link1](%s/fail1)
Line 2 OK text
Line 3 [link3](%s/fail3)
Line 4 OK text
Line 5 [link5](%s/fail5)
Line 6 [link6](%s/fail6)
Line 7 OK text
Line 8 [link8](%s/fail8)
`, ts.URL, ts.URL, ts.URL, ts.URL, ts.URL)

	skip := []*regexp.Regexp{}
	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("mock.md", lines, offset, skip, 10, 10, &sync.Map{})

	if len(results) != 5 {
		t.Fatalf("expected 5 errors, got %d", len(results))
	}

	// Verify all expected lines have errors (order doesn't matter)
	expectedLines := map[int]bool{1: true, 3: true, 5: true, 6: true, 8: true}
	for _, result := range results {
		if !expectedLines[result.Line] {
			t.Errorf("unexpected error at line %d", result.Line)
		}
		delete(expectedLines, result.Line)
	}
	if len(expectedLines) > 0 {
		t.Errorf("missing errors at lines: %v", expectedLines)
	}
}

func TestCheckExternalLinks_Deduplication(t *testing.T) {
	requestCount := 0
	customTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer customTs.Close()

	urlCache := &sync.Map{}
	markdown := fmt.Sprintf("[link](%s/fail)", customTs.URL)
	lines, offset := toLines(markdown)

	// First call - should trigger HTTP request
	rule.CheckExternalLinks("file1.md", lines, offset, []*regexp.Regexp{}, 10, 10, urlCache)
	// Second call - should use cache
	rule.CheckExternalLinks("file2.md", lines, offset, []*regexp.Regexp{}, 10, 10, urlCache)

	if requestCount != 1 {
		t.Errorf("expected only 1 HTTP request due to caching, but got %d", requestCount)
	}
}

func TestCheckExternalLinks_MultipleOccurrences(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	markdown := fmt.Sprintf("[fail](%s/fail)\n[fail again](%s/fail)", ts.URL, ts.URL)

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("mock.md", lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	// Should report errors for each occurrence
	if len(results) != 2 {
		t.Fatalf("expected 2 errors for repeated link, got %d", len(results))
	}
	// Verify both lines have errors (order may vary due to parallel processing)
	expectedLines := map[int]bool{1: true, 2: true}
	for _, res := range results {
		if !expectedLines[res.Line] {
			t.Errorf("unexpected error on line %d", res.Line)
		}
		delete(expectedLines, res.Line)
	}
	if len(expectedLines) > 0 {
		t.Errorf("missing errors at lines: %v", expectedLines)
	}
}

func TestCheckExternalLinks_AllLinksSucceed(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	markdown := fmt.Sprintf(`[link1](%s/ok)
[link2](%s/ok)
[link3](%s/ok)`, ts.URL, ts.URL, ts.URL)

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("mock.md", lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	if len(results) != 0 {
		t.Errorf("expected 0 errors for all successful links, got %d", len(results))
	}
}

func TestCheckExternalLinks_HTTPStatusBoundary(t *testing.T) {
	boundaryTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/399":
			w.WriteHeader(399)
		case "/400":
			w.WriteHeader(400)
		}
	}))
	defer boundaryTs.Close()

	markdown := fmt.Sprintf("[success](%s/399)\n[fail](%s/400)", boundaryTs.URL, boundaryTs.URL)
	fileName := "boundary.md"

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	if len(results) != 1 {
		t.Fatalf("expected 1 error (only 400), got %d", len(results))
	}

	got := results[0]
	if got.Line != 2 {
		t.Errorf("expected error at line 2 (400 status), got line %d", got.Line)
	}
	if got.File != fileName {
		t.Errorf("expected file %q, got %q", fileName, got.File)
	}
	if !strings.Contains(got.Message, boundaryTs.URL+"/400") {
		t.Errorf("message %q does not contain URL %q", got.Message, boundaryTs.URL+"/400")
	}
}

func TestCheckExternalLinks_MultipleSkipPatterns(t *testing.T) {
	markdown := `
[localhost](http://localhost/skip)
[example](http://example.com/skip)
[test](http://test.internal/skip)
[check](https://httpstat.us/404)
`
	fileName := "skip.md"
	skip := []*regexp.Regexp{
		regexp.MustCompile(`localhost`),
		regexp.MustCompile(`example\.com`),
		regexp.MustCompile(`\.internal`),
	}

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(fileName, lines, offset, skip, 2, 10, &sync.Map{})

	if len(results) != 1 {
		t.Fatalf("expected 1 error (only httpstat.us), got %d", len(results))
	}

	got := results[0]
	if got.Line != 5 {
		t.Errorf("expected error at line 5 (httpstat.us), got line %d", got.Line)
	}
	if got.File != fileName {
		t.Errorf("expected file %q, got %q", fileName, got.File)
	}
	if !strings.Contains(got.Message, "httpstat.us") {
		t.Errorf("message %q does not contain 'httpstat.us'", got.Message)
	}
}

func TestCheckExternalLinks_NetworkError(t *testing.T) {
	// Use an invalid/unreachable URL that will cause network error
	markdown := "[unreachable](http://invalid.test.localhost.invalid:9999/path)"
	fileName := "network.md"
	unreachableURL := "http://invalid.test.localhost.invalid:9999/path"

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 1, 10, &sync.Map{})

	if len(results) != 1 {
		t.Fatalf("expected 1 error for network failure, got %d", len(results))
	}

	got := results[0]
	if got.Line != 1 {
		t.Errorf("expected error at line 1, got line %d", got.Line)
	}
	if got.File != fileName {
		t.Errorf("expected file %q, got %q", fileName, got.File)
	}
	if !strings.Contains(got.Message, unreachableURL) {
		t.Errorf("message %q does not contain URL %q", got.Message, unreachableURL)
	}
}

func TestCheckExternalLinks_DifferentHTTPStatusCodes(t *testing.T) {
	statusTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/404":
			w.WriteHeader(404)
		case "/500":
			w.WriteHeader(500)
		case "/503":
			w.WriteHeader(503)
		}
	}))
	defer statusTs.Close()

	markdown := fmt.Sprintf("[not-found](%s/404)\n[server-error](%s/500)\n[unavailable](%s/503)", statusTs.URL, statusTs.URL, statusTs.URL)
	fileName := "test.md"
	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	if len(results) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(results))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Line < results[j].Line
	})

	expected := []struct {
		line int
		url  string
	}{
		{1, statusTs.URL + "/404"},
		{2, statusTs.URL + "/500"},
		{3, statusTs.URL + "/503"},
	}

	for i, exp := range expected {
		got := results[i]
		if got.Line != exp.line {
			t.Errorf("Result[%d]: expected line %d, got %d", i, exp.line, got.Line)
		}
		if !strings.Contains(got.Message, exp.url) {
			t.Errorf("Result[%d]: message %q does not contain URL %q", i, got.Message, exp.url)
		}
	}
}

func TestCheckExternalLinks_RetrySuccess(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	markdown := fmt.Sprintf("[retry-link](%s/retry)", ts.URL)

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("retry.md", lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	if len(results) != 0 {
		t.Errorf("expected 0 errors due to successful retry, but got %d", len(results))
	}

	if requestCount != 2 {
		t.Errorf("expected exactly 2 requests (1 fail + 1 retry), but got %d", requestCount)
	}
}

func TestCheckExternalLinks_NoRetryFor404(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	markdown := fmt.Sprintf("[not-found](%s/404)", ts.URL)

	lines, offset := toLines(markdown)
	rule.CheckExternalLinks("404.md", lines, offset, []*regexp.Regexp{}, 10, 10, &sync.Map{})

	// For 404, there should be no retry; it should give up after a single request
	if requestCount != 1 {
		t.Errorf("expected only 1 request for 404 (no retry), but got %d", requestCount)
	}
}

func TestCheckExternalLinks_ConcurrencyLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	var links []string
	for i := range 100 {
		links = append(links, fmt.Sprintf("[L%d](%s/%d)", i, ts.URL, i))
	}
	markdown := strings.Join(links, "\n")

	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks("heavy.md", lines, offset, []*regexp.Regexp{}, 5, 10, &sync.Map{})

	if len(results) != 0 {
		t.Errorf("expected 0 errors, got %d", len(results))
	}
}

func TestCheckExternalLinks_CacheErrorState(t *testing.T) {
	// Test that network errors are cached and reported consistently across multiple calls
	markdown := "[unreachable](http://invalid.test.localhost.invalid:9999/path)"
	fileName := "cache-error.md"
	urlCache := &sync.Map{}

	// First call - should trigger network error
	lines, offset := toLines(markdown)
	results1 := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 1, 10, urlCache)

	if len(results1) != 1 {
		t.Fatalf("first call: expected 1 error for network failure, got %d", len(results1))
	}

	// Second call with same cache - should use cached error and report the same error
	results2 := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 1, 10, urlCache)

	if len(results2) != 1 {
		t.Fatalf("second call: expected 1 error from cache, got %d", len(results2))
	}

	// Verify both calls produce the same error
	if results1[0].Message != results2[0].Message {
		t.Errorf("error messages differ: first=%q, second=%q", results1[0].Message, results2[0].Message)
	}
}

func TestCheckExternalLinks_CacheInvalidType(t *testing.T) {
	// Test that when cache contains an unexpected type, it re-checks the URL
	ts := setupTestServer()
	defer ts.Close()

	markdown := fmt.Sprintf("[link](%s/fail)", ts.URL)
	fileName := "invalid-cache.md"
	urlCache := &sync.Map{}

	// Pre-populate cache with invalid type (string instead of cacheResult)
	testURL := ts.URL + "/fail"
	urlCache.Store(testURL, "invalid type")

	// Call CheckExternalLinks - should detect invalid cache type and re-check
	lines, offset := toLines(markdown)
	results := rule.CheckExternalLinks(fileName, lines, offset, []*regexp.Regexp{}, 10, 10, urlCache)

	// Should still get error because URL returns 404
	if len(results) != 1 {
		t.Fatalf("expected 1 error after re-checking invalid cache, got %d", len(results))
	}

	// Verify that cache was updated with correct type
	if cached, ok := urlCache.Load(testURL); ok {
		// Just verify something is stored - we can't access the private type from test
		if cached == nil {
			t.Error("cache entry is nil after re-check")
		}
		// The fact that it didn't panic and produced correct results proves it worked
	} else {
		t.Error("cache entry was not found after re-check")
	}
}
