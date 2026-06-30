package rule_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"sync"
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
	"github.com/shinagawa-web/gomarklint/v3/internal/rule"
)

// TestCheckExternalLinks_NoRequestsForBlockedContexts is the #337 Phase 4 live
// verification: the static audit claimed external-link no longer issues HTTP
// requests for URLs that appear inside indented code, HTML blocks, HTML
// comments, or inline code spans. This confirms it against a real run by
// recording every path the server is asked for and asserting only the control
// link (outside any such context) is ever fetched.
//
// Unlike the results-based TestCheckExternalLinks_IgnoreCodeBlocks, this asserts
// on the actual network traffic, so it cannot be fooled by a blocked URL that
// happens to return 200 (which would also produce zero lint errors).
func TestCheckExternalLinks_NoRequestsForBlockedContexts(t *testing.T) {
	var mu sync.Mutex
	requested := make(map[string]int)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requested[r.URL.Path]++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// One URL per blocked context plus a control link outside every context.
	// The blocked contexts are separated by blank lines so the scanner classifies
	// them as their own blocks (indented code, HTML block, HTML comment).
	markdown := fmt.Sprintf(`[control](%[1]s/control)

`+"`[inline](%[1]s/inline-code)`"+` after an inline code span

    [indented](%[1]s/indented)

<div>
[html-block](%[1]s/html-block)
</div>

<!-- [html-comment](%[1]s/html-comment) -->
`, ts.URL)

	lines, offset := toLines(markdown)
	results, checked := rule.CheckExternalLinks(
		"mock.md", preprocess.Scan(lines), offset,
		[]*regexp.Regexp{}, 10, 10,
		rule.DefaultMaxConcurrency, rule.DefaultMaxRetries,
		nil, &sync.Map{}, 0, 0,
	)

	// Only the control URL is a real link, and it returns 200, so no errors.
	if len(results) != 0 {
		t.Fatalf("expected 0 lint errors, got %d: %+v", len(results), results)
	}
	// Exactly one unique URL should have been extracted and checked.
	if checked != 1 {
		t.Errorf("expected 1 checked URL (control only), got %d", checked)
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := requested["/control"]; !ok {
		t.Errorf("expected the control link to be fetched, but it was not; requested paths: %v", keys(requested))
	}

	blocked := []string{"/inline-code", "/indented", "/html-block", "/html-comment"}
	for _, path := range blocked {
		if n, ok := requested[path]; ok {
			t.Errorf("URL inside a blocked context was fetched %d time(s): %s", n, path)
		}
	}

	if got := keys(requested); len(got) != 1 {
		t.Errorf("expected exactly one requested path (/control), got %v", got)
	}
}

func keys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
