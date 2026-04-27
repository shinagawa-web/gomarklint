package rule

import (
	"strings"
	"testing"
)

func TestCheckLinkFragments(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		opts     map[string]interface{}
		wantErrs []LintError
	}{
		{
			name:     "valid: no fragment links",
			content:  "## Hello\n\nSome text.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: fragment link matches heading",
			content:  "## Introduction\n\nSee [Introduction](#introduction) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: fragment link with hyphenated heading",
			content:  "## Getting Started\n\nSee [Getting Started](#getting-started) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:    "invalid: fragment link not found",
			content: "## Introduction\n\nSee [Setup](#setup) for details.\n",
			opts:    map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "link-fragments: fragment #setup not found in this document"},
			},
		},
		{
			name:     "valid: first duplicate heading uses bare slug",
			content:  "## Intro\n\nSee [First Intro](#intro) for details.\n\n## Intro\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: second duplicate heading uses suffixed slug",
			content:  "## Intro\n\n## Intro\n\nSee [Second Intro](#intro-1) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: heading inside fenced code block is excluded",
			content:  "## Real Heading\n\n```\n## Fake Heading\n```\n\nSee [real](#real-heading).\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:    "invalid: heading inside fenced code block does not produce a slug",
			content: "```\n## Fake Heading\n```\n\nSee [fake](#fake-heading).\n",
			opts:    map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: []LintError{
				{File: "test.md", Line: 5, Message: "link-fragments: fragment #fake-heading not found in this document"},
			},
		},
		{
			name:     "valid: fragment link inside fenced code block is ignored",
			content:  "## Hello\n\n```\nSee [broken](#broken-link) here.\n```\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: fragment link inside inline code is ignored",
			content:  "## Hello\n\nUse `[broken](#broken-link)` here.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: reference link with matching fragment definition",
			content:  "## Introduction\n\nSee [Intro][intro-ref] for details.\n\n[intro-ref]: #introduction\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:    "invalid: reference link with non-existent fragment",
			content: "## Introduction\n\nSee [Setup][setup-ref] for details.\n\n[setup-ref]: #setup\n",
			opts:    map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "link-fragments: fragment #setup not found in this document"},
			},
		},
		{
			name:     "valid: reference link pointing to external URL is ignored",
			content:  "## Hello\n\nSee [Example][ex].\n\n[ex]: https://example.com\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: non-fragment inline links are ignored",
			content:  "## Hello\n\nSee [Example](https://example.com) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: heading with inline code slug",
			content:  "## The `go test` Command\n\nSee [go test](#the-go-test-command) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: heading with bold formatting",
			content:  "## **Introduction**\n\nSee [Introduction](#introduction) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: nil,
		},
		{
			name:     "valid: gitlab algorithm",
			content:  "## Hello World\n\nSee [Hello](#hello-world) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "gitlab"},
			wantErrs: nil,
		},
		{
			name:     "valid: default algorithm is github when option absent",
			content:  "## Hello World\n\nSee [Hello](#hello-world) for details.\n",
			opts:     map[string]interface{}{},
			wantErrs: nil,
		},
		{
			name:     "valid: unknown algorithm falls back to github",
			content:  "## Hello World\n\nSee [Hello](#hello-world) for details.\n",
			opts:     map[string]interface{}{"slug-algorithm": "unknown-algo"},
			wantErrs: nil,
		},
		{
			name:    "invalid: multiple broken fragment links on different lines",
			content: "## Hello\n\nSee [broken1](#broken1) and\nSee [broken2](#broken2) too.\n",
			opts:    map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: []LintError{
				{File: "test.md", Line: 3, Message: "link-fragments: fragment #broken1 not found in this document"},
				{File: "test.md", Line: 4, Message: "link-fragments: fragment #broken2 not found in this document"},
			},
		},
		{
			name:    "valid: offset applied to line numbers",
			content: "## Introduction\n\nSee [broken](#broken).\n",
			opts:    map[string]interface{}{"slug-algorithm": "github"},
			wantErrs: []LintError{
				{File: "test.md", Line: 8, Message: "link-fragments: fragment #broken not found in this document"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			offset := 0
			// The "offset applied" test uses offset=5
			if tt.name == "valid: offset applied to line numbers" {
				offset = 5
			}
			got := CheckLinkFragments("test.md", lines, offset, tt.opts)

			if len(got) != len(tt.wantErrs) {
				t.Fatalf("got %d errors, want %d:\n  got:  %v\n  want: %v", len(got), len(tt.wantErrs), got, tt.wantErrs)
			}
			for i, g := range got {
				w := tt.wantErrs[i]
				if g.File != w.File || g.Line != w.Line || g.Message != w.Message {
					t.Errorf("error[%d]:\n  got  {%s:%d %q}\n  want {%s:%d %q}",
						i, g.File, g.Line, g.Message, w.File, w.Line, w.Message)
				}
			}
		})
	}
}

func TestExtractHeadingText(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantText  string
		wantLevel int
	}{
		{"h1", "# Hello", "Hello", 1},
		{"h2", "## Hello", "Hello", 2},
		{"h3", "### Hello World", "Hello World", 3},
		{"h6", "###### Hello", "Hello", 6},
		{"h7 too deep", "####### Hello", "", 0},
		{"no space after #", "##nospace", "", 0},
		{"not a heading", "some text", "", 0},
		{"empty heading", "## ", "", 2},
		{"heading with extra spaces trimmed", "##   Hello  ", "Hello", 2},
		{"heading with inline formatting kept", "## **Hello** World", "**Hello** World", 2},
		{"empty string", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotLevel := extractHeadingText(tt.line)
			if gotText != tt.wantText || gotLevel != tt.wantLevel {
				t.Errorf("extractHeadingText(%q) = (%q, %d), want (%q, %d)",
					tt.line, gotText, gotLevel, tt.wantText, tt.wantLevel)
			}
		})
	}
}

func TestCollectRefDefs(t *testing.T) {
	t.Run("collects fragment refs", func(t *testing.T) {
		lines := strings.Split("[ref1]: #section-1\n[ref2]: #section-2\n[ext]: https://example.com\n", "\n")
		defs := collectRefDefs(lines)
		if defs["ref1"] != "section-1" {
			t.Errorf("expected ref1 -> section-1, got %q", defs["ref1"])
		}
		if defs["ref2"] != "section-2" {
			t.Errorf("expected ref2 -> section-2, got %q", defs["ref2"])
		}
		if _, ok := defs["ext"]; ok {
			t.Error("external link ref should not be collected")
		}
	})

	t.Run("label is normalized to lowercase", func(t *testing.T) {
		lines := []string{"[My Label]: #my-section"}
		defs := collectRefDefs(lines)
		if defs["my label"] != "my-section" {
			t.Errorf("expected 'my label' -> 'my-section', got %q", defs["my label"])
		}
	})
}
