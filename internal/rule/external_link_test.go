package rule_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/rule"
)

func TestCheckExternalLinks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/fail":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	markdown := fmt.Sprintf(`
[working](%s/ok)
[broken](%s/fail)
`, ts.URL, ts.URL)

	results := rule.CheckExternalLinks("mock.md", markdown)

	if len(results) != 1 {
		t.Fatalf("expected 1 error, got %d", len(results))
	}

	got := results[0]
	if !strings.Contains(got.Message, "/fail") || !strings.Contains(got.Message, "404") {
		t.Errorf("unexpected error message: %s", got.Message)
	}
}
