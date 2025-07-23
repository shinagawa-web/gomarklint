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

	markdown := fmt.Sprintf(`# Title

This link should pass: [working](%s/ok)

This link should fail: [broken](%s/fail)
`, ts.URL, ts.URL)

	results := rule.CheckExternalLinks("mock.md", markdown)

	if len(results) != 1 {
		t.Fatalf("expected 1 error, got %d", len(results))
	}

	got := results[0]

	if !strings.Contains(got.Message, "/fail") || !strings.Contains(got.Message, "404") {
		t.Errorf("unexpected error message: %s", got.Message)
	}

	if got.Line != 5 {
		t.Errorf("expected line number 5, got %d", got.Line)
	}

	if got.File != "mock.md" {
		t.Errorf("expected file 'mock.md', got %s", got.File)
	}
}
