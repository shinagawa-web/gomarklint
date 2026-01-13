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
		results := rule.CheckExternalLinks(file, markdown, []*regexp.Regexp{}, 10)

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
	})

	t.Run("skip pattern should exclude localhost", func(t *testing.T) {
		markdown := `
[skip this](http://localhost/skip)
[check this](https://httpstat.us/404)
`
		skip := []*regexp.Regexp{
			regexp.MustCompile(`localhost`),
		}

		results := rule.CheckExternalLinks("mock.md", markdown, skip, 10)

		if len(results) != 1 {
			t.Fatalf("expected 1 error (only non-localhost link should be checked), got %d", len(results))
		}
		if !strings.Contains(results[0].Message, "httpstat.us") {
			t.Errorf("expected error for httpstat.us link, got: %v", results[0])
		}
	})
	t.Run("ignore links inside code blocks", func(t *testing.T) {
		markdown := fmt.Sprintf("```\n[code link](%s/in-code)\n```\n[real link](%s/fail)\n", ts.URL, ts.URL)

		skip := []*regexp.Regexp{}
		results := rule.CheckExternalLinks("mock.md", markdown, skip, 10)
		if len(results) != 1 {
			t.Fatalf("expected 1 error (code block link should be ignored), got %d", len(results))
		}
		if !strings.Contains(results[0].Message, "/fail") {
			t.Errorf("expected error for real link, got: %v", results[0])
		}
		if results[0].Line != 4 {
			t.Errorf("expected line 4 for real link, got: %d", results[0].Line)
		}
	})

	t.Run("parallel check returns errors in line number order", func(t *testing.T) {
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
		results := rule.CheckExternalLinks("mock.md", markdown, skip, 10)

		if len(results) != 5 {
			t.Fatalf("expected 5 errors, got %d", len(results))
		}

		// Verify errors are returned in ascending line number order
		expectedLines := []int{1, 3, 5, 6, 8}
		for i, result := range results {
			if result.Line != expectedLines[i] {
				t.Errorf("expected error at line %d, got line %d", expectedLines[i], result.Line)
			}
		}
	})
}
