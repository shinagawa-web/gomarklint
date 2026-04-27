package rule

import (
	"testing"
)

func TestComputeSlug(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		algorithm string
		want      string
	}{
		// GitHub (default)
		{"github: simple text", "Hello World", "github", "hello-world"},
		{"github: comma and exclamation stripped", "Hello, World!", "github", "hello-world"},
		{"github: period stripped", "Subsection 1.1", "github", "subsection-11"},
		{"github: consecutive spaces become consecutive hyphens", "Hello  World", "github", "hello--world"},
		{"github: preserves Unicode letters", "日本語", "github", "日本語"},
		{"github: hyphen preserved", "go-lang", "github", "go-lang"},
		{"github: underscore preserved", "go_lang", "github", "go_lang"},
		{"github: hash stripped", "C#", "github", "c"},
		{"github: plus stripped", "C++", "github", "c"},
		{"github: colon stripped", "foo: bar", "github", "foo-bar"},
		{"github: section number", "Section 1", "github", "section-1"},

		// GitLab (goldmark)
		{"gitlab: simple text", "Hello World", "gitlab", "hello-world"},
		{"gitlab: period stripped, collapses", "Subsection 1.1", "gitlab", "subsection-11"},
		{"gitlab: preserves Unicode", "日本語", "gitlab", "日本語"},
		{"gitlab: consecutive hyphens collapsed", "A  B", "gitlab", "a-b"},
		{"gitlab: punctuation stripped", "Hello, World!", "gitlab", "hello-world"},

		// Zenn (markdown-it-anchor)
		{"zenn: simple text", "Hello World", "zenn", "hello-world"},
		{"zenn: preserves punctuation", "Hello, World!", "zenn", "hello,-world!"},
		{"zenn: preserves Unicode", "日本語", "zenn", "日本語"},
		{"zenn: period preserved", "Go 1.21", "zenn", "go-1.21"},

		// Pandoc (auto_identifiers)
		{"pandoc: simple text", "Hello World", "pandoc", "hello-world"},
		{"pandoc: strips non-ASCII", "日本語", "pandoc", ""},
		{"pandoc: period stripped", "Go 1.21", "pandoc", "go-121"},
		{"pandoc: collapses hyphens from double space", "A  B", "pandoc", "a-b"},

		// Kramdown
		{"kramdown: simple text", "Hello World", "kramdown", "hello-world"},
		{"kramdown: strips non-ASCII", "日本語", "kramdown", ""},
		{"kramdown: period stripped", "Go 1.21", "kramdown", "go-121"},
		{"kramdown: underscore stripped (not in keep set)", "go_lang", "kramdown", "golang"},

		// Hugo (= GitHub)
		{"hugo: same as github", "Hello World", "hugo", "hello-world"},
		{"hugo: period stripped", "Go 1.21", "hugo", "go-121"},

		// MkDocs
		{"mkdocs: simple text", "Hello World", "mkdocs", "hello-world"},
		{"mkdocs: strips non-ASCII", "日本語", "mkdocs", ""},
		{"mkdocs: collapses hyphens", "A  B", "mkdocs", "a-b"},
		{"mkdocs: period stripped", "Go 1.21", "mkdocs", "go-121"},

		// DocFX (case-preserving)
		{"docfx: preserves case", "Hello World", "docfx", "Hello-World"},
		{"docfx: preserves period", "Subsection 1.1", "docfx", "Subsection-1.1"},
		{"docfx: strips non-ASCII", "日本語", "docfx", ""},
		{"docfx: strips punctuation except hyphen underscore dot", "Hello, World!", "docfx", "Hello-World"},

		// Aliases
		{"pandoc-gfm: same as github", "Hello World", "pandoc-gfm", "hello-world"},

		// Unknown falls back to github
		{"unknown algorithm: falls back to github", "Hello World", "unknown-algo", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeSlug(tt.text, tt.algorithm)
			if got != tt.want {
				t.Errorf("ComputeSlug(%q, %q) = %q, want %q", tt.text, tt.algorithm, got, tt.want)
			}
		})
	}
}

func TestStripHeadingFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text unchanged", "Hello World", "Hello World"},
		{"bold asterisk stripped", "**Hello** World", "Hello World"},
		{"bold underscore stripped", "__Hello__ World", "Hello World"},
		{"italic asterisk stripped", "*Hello* World", "Hello World"},
		{"italic underscore stripped", "_Hello_ World", "Hello World"},
		{"inline code stripped (backtick delimiters removed)", "`code` example", "code example"},
		{"link text kept, URL removed", "[Hello](https://example.com)", "Hello"},
		{"image alt text kept, URL removed", "![alt text](image.png)", "alt text"},
		{"HTML tag removed", "<em>Hello</em> World", "Hello World"},
		{"ref link text kept", "[Hello][ref]", "Hello"},
		{"ref image alt kept", "![alt][img-ref]", "alt"},
		{"HTML comment removed", "<!-- note --> Hello", " Hello"},
		{"multi-backtick code", "``go test``", "go test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHeadingFormatting(tt.input)
			if got != tt.want {
				t.Errorf("stripHeadingFormatting(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCollapseDashes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello-world", "hello-world"},
		{"hello--world", "hello-world"},
		{"--hello--", "hello"},
		{"a---b", "a-b"},
		{"-leading", "leading"},
		{"trailing-", "trailing"},
		{"", ""},
		{"a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := collapseDashes(tt.input)
			if got != tt.want {
				t.Errorf("collapseDashes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildSlugSet(t *testing.T) {
	t.Run("unique headings", func(t *testing.T) {
		headings := []string{"Introduction", "Getting Started", "References"}
		slugs := buildSlugSet(headings, "github")
		for _, want := range []string{"introduction", "getting-started", "references"} {
			if _, ok := slugs[want]; !ok {
				t.Errorf("expected slug %q to be in set", want)
			}
		}
	})

	t.Run("duplicate headings generate suffixed slugs", func(t *testing.T) {
		headings := []string{"Intro", "Intro", "Intro"}
		slugs := buildSlugSet(headings, "github")
		for _, want := range []string{"intro", "intro-1", "intro-2"} {
			if _, ok := slugs[want]; !ok {
				t.Errorf("expected slug %q to be in set", want)
			}
		}
		if _, ok := slugs["intro-0"]; ok {
			t.Error("slug intro-0 should not exist (first occurrence stays bare)")
		}
	})

	t.Run("empty heading text produces no slug", func(t *testing.T) {
		headings := []string{"", "Hello"}
		slugs := buildSlugSet(headings, "github")
		if len(slugs) != 1 {
			t.Errorf("expected 1 slug, got %d: %v", len(slugs), slugs)
		}
		if _, ok := slugs["hello"]; !ok {
			t.Error("expected slug 'hello'")
		}
	})
}
