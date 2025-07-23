package rule_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
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

	t.Run("basic success/failure", func(t *testing.T) {
		markdown := fmt.Sprintf(`[ok](%s/ok)
[fail](%s/fail)
`, ts.URL, ts.URL)

		file := "mock.md"
		results := rule.CheckExternalLinks(file, markdown, []*regexp.Regexp{})

		if len(results) != 1 {
			t.Fatalf("expected 1 error, got %d", len(results))
		}

		got := results[0]
		if !strings.Contains(got.Message, "/fail") || !strings.Contains(got.Message, "404") {
			t.Errorf("unexpected error message: %s", got.Message)
		}
		if got.Line != 2 {
			t.Errorf("expected line number 2, got %d", got.Line)
		}
		if got.File != file {
			t.Errorf("expected file '%s', got %s", file, got.File)
		}
	})

	t.Run("skip pattern should exclude localhost", func(t *testing.T) {
		markdown := `
[skip this](http://localhost/skip)
[check this](https://httpstat.us/404)
`
		skip := []*regexp.Regexp{
			regexp.MustCompile(`localhost`),
		}

		results := rule.CheckExternalLinks("mock.md", markdown, skip)

		if len(results) != 1 {
			t.Fatalf("expected 1 error (only non-localhost link should be checked), got %d", len(results))
		}
		if !strings.Contains(results[0].Message, "httpstat.us") {
			t.Errorf("expected error for httpstat.us link, got: %v", results[0])
		}
	})

}
