package linter

import (
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/config"
)

func FuzzLintContent(f *testing.F) {
	seeds := []string{
		"# Hello\n\nWorld\n",
		"## Section\n\nSome text with **bold** and _italic_.\n",
		"```go\nfmt.Println(\"hello\")\n```\n",
		"- item one\n- item two\n- item three\n",
		"[link text](https://example.com)\n",
		"![alt text](image.png)\n",
		"| col1 | col2 |\n|------|------|\n| a    | b    |\n",
		"    code with hard tab\n",
		"Line that is intentionally very long and exceeds the default maximum line length limit set by the linter configuration.\n",
		"",
		"---\ntitle: frontmatter\n---\n\n# Body\n",
		"# H1\n\n## H2\n\n### H3\n\n#### H4\n",
		"https://bare-url.example.com\n",
		"> blockquote\n",
		"* [ ] task\n* [x] done\n",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	cfg := config.Default()
	// Disable external-link to avoid network calls during fuzzing.
	cfg.Rules["external-link"] = &config.RuleConfig{
		Enabled:  false,
		Severity: config.SeverityOff,
		Options:  map[string]interface{}{},
	}

	l, err := New(cfg)
	if err != nil {
		f.Fatalf("failed to create linter: %v", err)
	}

	f.Fuzz(func(t *testing.T, data string) {
		l.LintContent("fuzz.md", data)
	})
}
