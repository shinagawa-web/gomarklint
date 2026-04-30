package rule

import (
	"regexp"
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
		{"github: em dash (U+2014, General Punctuation) stripped", "hello—world", "github", "helloworld"},
		{"github: supplemental punctuation (U+2E3A) stripped", "hello⸺world", "github", "helloworld"},
		{"github: emoji stripped (globe)", "Hello 🌍", "github", "hello-"},
		{"github: emoji stripped (rocket at start)", "🚀 Getting Started", "github", "-getting-started"},
		{"github: accented Latin preserved", "café", "github", "café"},

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
		{"pandoc: period kept", "Go 1.21", "pandoc", "go-1.21"},
		{"pandoc: collapses hyphens from double space", "A  B", "pandoc", "a-b"},
		{"pandoc: leading digits stripped", "123abc", "pandoc", "abc"},
		{"pandoc: numbered list item strips prefix", "1. Introduction", "pandoc", "introduction"},
		{"pandoc: section with period", "Section 1.1", "pandoc", "section-1.1"},

		// Kramdown
		{"kramdown: simple text", "Hello World", "kramdown", "hello-world"},
		{"kramdown: strips non-ASCII", "日本語", "kramdown", ""},
		{"kramdown: period stripped", "Go 1.21", "kramdown", "go-121"},
		{"kramdown: underscore stripped (not in keep set)", "go_lang", "kramdown", "golang"},
		{"kramdown: leading digits stripped", "123abc", "kramdown", "abc"},
		{"kramdown: single leading digit stripped", "1test", "kramdown", "test"},

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

		// Qiita / mdbook (same character set)
		{"qiita: simple text", "Hello World", "qiita", "hello-world"},
		{"qiita: preserves Unicode letters", "日本語", "qiita", "日本語"},
		{"qiita: preserves underscore", "go_lang", "qiita", "go_lang"},
		{"qiita: strips punctuation", "Hello, World!", "qiita", "hello-world"},
		{"qiita: no collapse", "A  B", "qiita", "a--b"},
		{"mdbook: same as qiita", "Hello World", "mdbook", "hello-world"},
		{"mdbook: preserves Unicode", "日本語", "mdbook", "日本語"},

		// VitePress
		{"vitepress: simple text", "Hello World", "vitepress", "hello-world"},
		{"vitepress: accented Latin normalized", "Héllo Wörld", "vitepress", "hello-world"},
		{"vitepress: preserves CJK", "日本語", "vitepress", "日本語"},
		{"vitepress: collapses non-alphanumeric", "A  B", "vitepress", "a-b"},

		// Gitea / Forgejo
		{"gitea: prefixes user-content-", "Hello World", "gitea", "user-content-hello-world"},
		{"gitea: preserves Unicode", "日本語", "gitea", "user-content-日本語"},
		{"forgejo: same as gitea", "Hello World", "forgejo", "user-content-hello-world"},

		// Sphinx
		{"sphinx: simple text", "Hello World", "sphinx", "hello-world"},
		{"sphinx: accented Latin normalized", "Héllo", "sphinx", "hello"},
		{"sphinx: strips non-ASCII CJK", "日本語", "sphinx", ""},
		{"sphinx: collapses non-alphanumeric", "A  B", "sphinx", "a-b"},
		{"sphinx: punctuation collapsed", "Go 1.21", "sphinx", "go-1-21"},
		{"sphinx: strips leading digits", "123abc", "sphinx", "abc"},
		{"sphinx: strips single leading digit", "1test", "sphinx", "test"},
		{"sphinx: preserves interior digits", "abc123def", "sphinx", "abc123def"},
		{"sphinx: all digits produces empty", "123", "sphinx", ""},

		// Eleventy
		{"eleventy: simple text", "Hello World", "eleventy", "hello-world"},
		{"eleventy: accented Latin normalized", "Héllo", "eleventy", "hello"},
		{"eleventy: strips CJK", "日本語", "eleventy", ""},
		{"eleventy: collapses hyphens", "A  B", "eleventy", "a-b"},
		{"eleventy: preserves hyphen", "go-lang", "eleventy", "go-lang"},
		{"eleventy: period becomes hyphen", "Section 1.1", "eleventy", "section-1-1"},
		{"eleventy: underscore becomes hyphen", "foo_bar", "eleventy", "foo-bar"},
		{"eleventy: umlaut ü expanded", "über", "eleventy", "ueber"},
		{"eleventy: umlaut Ö expanded", "Österreich", "eleventy", "oesterreich"},
		{"eleventy: CJK between ASCII becomes separator", "A日本B", "eleventy", "a-b"},

		// Azure DevOps
		{"azure-devops: simple text", "Hello World", "azure-devops", "hello-world"},
		{"azure-devops: special char encoded", "C# Tutorial", "azure-devops", "c%23-tutorial"},
		{"azure-devops: CJK percent-encoded", "はじめに", "azure-devops", "%E3%81%AF%E3%81%98%E3%82%81%E3%81%AB"},
		{"azure-devops: RFC3986 unreserved kept", "go-1.0_test~", "azure-devops", "go-1.0_test~"},

		// MyST
		{"myst: same as github", "Hello World", "myst", "hello-world"},

		// Aliases (→ github)
		{"pandoc-gfm: same as github", "Hello World", "pandoc-gfm", "hello-world"},
		{"docusaurus: same as github", "Hello World", "docusaurus", "hello-world"},
		{"gatsby: same as github", "Hello World", "gatsby", "hello-world"},
		{"astro: same as github", "Hello World", "astro", "hello-world"},
		{"starlight: same as github", "Hello World", "starlight", "hello-world"},
		{"nuxt-content: same as github", "Hello World", "nuxt-content", "hello-world"},

		// Alias → pandoc
		{"quarto: same as pandoc", "Hello World", "quarto", "hello-world"},
		{"quarto: strips non-ASCII", "日本語", "quarto", ""},

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
		{"code span with underscores not split by italic", "`my_func_name`", "my_func_name"},
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

func githubSlugger(text string) string { return ComputeSlug(text, "github") }

func TestBuildSlugSet(t *testing.T) {
	t.Run("unique headings", func(t *testing.T) {
		headings := []string{"Introduction", "Getting Started", "References"}
		slugs := buildSlugSet(headings, githubSlugger)
		for _, want := range []string{"introduction", "getting-started", "references"} {
			if _, ok := slugs[want]; !ok {
				t.Errorf("expected slug %q to be in set", want)
			}
		}
	})

	t.Run("duplicate headings generate suffixed slugs", func(t *testing.T) {
		headings := []string{"Intro", "Intro", "Intro"}
		slugs := buildSlugSet(headings, githubSlugger)
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
		slugs := buildSlugSet(headings, githubSlugger)
		if len(slugs) != 1 {
			t.Errorf("expected 1 slug, got %d: %v", len(slugs), slugs)
		}
		if _, ok := slugs["hello"]; !ok {
			t.Error("expected slug 'hello'")
		}
	})
}

func TestSlugCustom(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		params slugCustomParams
		want   string
	}{
		{
			name: "basic: lowercase + hyphen replacement",
			text: "Hello World",
			params: slugCustomParams{
				lowercase:        true,
				preserveUnicode:  true,
				spaceReplacement: '-',
			},
			want: "hello-world",
		},
		{
			name: "strip-chars removes matched characters",
			text: "Hello, World!",
			params: slugCustomParams{
				lowercase:        true,
				preserveUnicode:  true,
				spaceReplacement: '-',
				stripChars:       mustCompileRegexp(`[^\w\- ]`),
			},
			want: "hello-world",
		},
		{
			name: "preserve-unicode false strips non-ASCII",
			text: "Hello 日本語",
			params: slugCustomParams{
				lowercase:        true,
				preserveUnicode:  false,
				spaceReplacement: '-',
			},
			want: "hello-",
		},
		{
			name: "preserve-unicode true keeps non-ASCII",
			text: "Hello 日本語",
			params: slugCustomParams{
				lowercase:        true,
				preserveUnicode:  true,
				spaceReplacement: '-',
			},
			want: "hello-日本語",
		},
		{
			name: "collapse-separators collapses and trims",
			text: "  Hello  World  ",
			params: slugCustomParams{
				lowercase:          true,
				preserveUnicode:    true,
				spaceReplacement:   '-',
				collapseSeparators: true,
				collapseRe:         mustCompileRegexp("-+"),
			},
			want: "hello-world",
		},
		{
			name: "underscore space-replacement",
			text: "Hello World",
			params: slugCustomParams{
				lowercase:        true,
				preserveUnicode:  true,
				spaceReplacement: '_',
			},
			want: "hello_world",
		},
		{
			name: "no lowercase preserves case",
			text: "Hello World",
			params: slugCustomParams{
				lowercase:        false,
				preserveUnicode:  true,
				spaceReplacement: '-',
			},
			want: "Hello-World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugCustom(tt.text, tt.params)
			if got != tt.want {
				t.Errorf("slugCustom(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestMakeSlugger(t *testing.T) {
	t.Run("preset algorithm returns ComputeSlug result", func(t *testing.T) {
		slugger := makeSlugger("github", nil)
		got := slugger("Hello World")
		want := ComputeSlug("Hello World", "github")
		if got != want {
			t.Errorf("makeSlugger(github)(%q) = %q, want %q", "Hello World", got, want)
		}
	})

	t.Run("custom algorithm uses slug-params", func(t *testing.T) {
		opts := map[string]interface{}{
			"slug-algorithm": "custom",
			"slug-params": map[string]interface{}{
				"lowercase":           true,
				"preserve-unicode":    true,
				"space-replacement":   "-",
				"strip-chars":         `[^\w\- ]`,
				"collapse-separators": true,
			},
		}
		slugger := makeSlugger("custom", opts)
		got := slugger("Hello, World!")
		if got != "hello-world" {
			t.Errorf("custom slugger(%q) = %q, want %q", "Hello, World!", got, "hello-world")
		}
	})
}

func TestParseSlugParams(t *testing.T) {
	t.Run("invalid strip-chars regex is silently ignored", func(t *testing.T) {
		opts := map[string]interface{}{
			"slug-params": map[string]interface{}{
				"strip-chars": "[invalid(regex",
			},
		}
		p := parseSlugParams(opts)
		if p.stripChars != nil {
			t.Error("expected stripChars to be nil for invalid regex, got non-nil")
		}
	})

	t.Run("nil opts returns defaults", func(t *testing.T) {
		p := parseSlugParams(nil)
		if !p.lowercase || !p.preserveUnicode || p.spaceReplacement != '-' || p.collapseSeparators {
			t.Errorf("unexpected defaults: %+v", p)
		}
	})
}

func mustCompileRegexp(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
