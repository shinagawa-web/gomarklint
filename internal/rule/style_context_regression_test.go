package rule

import (
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func scanStyleDoc(doc string) *preprocess.Context {
	doc = strings.TrimPrefix(doc, "\n")
	return preprocess.Scan(strings.Split(doc, "\n"))
}

// TestStyleFamilySkipsBlockContexts covers the #337 Phase 3 (style family)
// gap closures for the five rules that skip all block contexts.
func TestStyleFamilySkipsBlockContexts(t *testing.T) {
	t.Run("consistent-list-marker", func(t *testing.T) {
		// A different marker inside indented code / HTML must not be flagged in
		// "dash" mode.
		for name, doc := range map[string]string{
			"indented":   "- real\n\n    * fake",
			"html block": "- real\n\n<div>\n* fake\n</div>",
		} {
			if errs := CheckConsistentListMarker("t.md", scanStyleDoc(doc), 0, "dash"); len(errs) != 0 {
				t.Errorf("%s: got %d, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("consistent-emphasis-style", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "*real*\n\n    _fake_",
			"inline code":  "*real*\n\nsee `_fake_` text",
			"html comment": "*real*\n\n<!-- _fake_ -->",
			"html block":   "*real*\n\n<div>\n_fake_\n</div>",
		} {
			// "asterisk" mode: an underscore span outside any block would be
			// flagged; inside these contexts it must not be.
			errs := CheckConsistentEmphasisStyle("t.md", scanStyleDoc(doc), 0, "asterisk")
			if len(errs) != 0 {
				t.Errorf("%s: got %d, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("blanks-around-lists", func(t *testing.T) {
		// A list-looking line inside indented code / HTML block must not drive
		// blank-line checks.
		for name, doc := range map[string]string{
			"indented":   "text\n\n    - item\n    - item2\n\nmore",
			"html block": "<div>\n- item\n- item2\n</div>",
		} {
			if errs := CheckBlanksAroundLists("t.md", scanStyleDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("no-multiple-blank-lines skips blanks inside a <pre> block", func(t *testing.T) {
		// A type-1 HTML block (<pre>) may legitimately contain consecutive blank
		// lines; they must not be flagged.
		doc := "<pre>\nline1\n\n\nline2\n</pre>"
		if errs := CheckNoMultipleBlankLines("t.md", scanStyleDoc(doc), 0); len(errs) != 0 {
			t.Errorf("got %d, want 0: %+v", len(errs), errs)
		}
		// Sanity: consecutive blanks in prose are still flagged.
		if errs := CheckNoMultipleBlankLines("t.md", scanStyleDoc("a\n\n\nb"), 0); len(errs) != 1 {
			t.Errorf("prose double blank: got %d, want 1", len(errs))
		}
	})

	t.Run("no-trailing-punctuation", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "text\n\n    # Heading:",
			"html comment": "<!-- # Heading: -->",
			"html block":   "<div>\n# Heading:\n</div>",
		} {
			if errs := CheckNoTrailingPunctuation("t.md", scanStyleDoc(doc), 0, ".,;:!?"); len(errs) != 0 {
				t.Errorf("%s: got %d, want 0: %+v", name, len(errs), errs)
			}
		}
	})
}

// TestSectionBStillChecksOutsideFencedCode is the regression guard for the
// markdownlint divergences (#337 Section B): max-line-length and no-hard-tabs
// intentionally skip ONLY fenced code, so they must still fire inside indented
// code and HTML blocks. These positive assertions fail if either rule is ever
// "tidied" to use the shared all-block skip.
func TestSectionBStillChecksOutsideFencedCode(t *testing.T) {
	long := strings.Repeat("x", 50)

	t.Run("max-line-length still flags long lines in indented code", func(t *testing.T) {
		doc := "text\n\n    " + long
		if errs := CheckMaxLineLength("t.md", scanStyleDoc(doc), 0, 40); len(errs) != 1 {
			t.Errorf("indented: got %d, want 1: %+v", len(errs), errs)
		}
	})

	t.Run("max-line-length still flags long lines in an HTML block", func(t *testing.T) {
		doc := "<div>\n" + long + "\n</div>"
		if errs := CheckMaxLineLength("t.md", scanStyleDoc(doc), 0, 40); len(errs) != 1 {
			t.Errorf("html block: got %d, want 1: %+v", len(errs), errs)
		}
	})

	t.Run("max-line-length still skips fenced code", func(t *testing.T) {
		doc := "```\n" + long + "\n```"
		if errs := CheckMaxLineLength("t.md", scanStyleDoc(doc), 0, 40); len(errs) != 0 {
			t.Errorf("fenced: got %d, want 0: %+v", len(errs), errs)
		}
	})

	t.Run("no-hard-tabs still flags tabs in indented code", func(t *testing.T) {
		// A tab-indented line is indented code; the tab must still be reported.
		doc := "text\n\n\tcode"
		if errs := CheckNoHardTabs("t.md", scanStyleDoc(doc), 0); len(errs) != 1 {
			t.Errorf("indented: got %d, want 1: %+v", len(errs), errs)
		}
	})

	t.Run("no-hard-tabs still skips fenced code", func(t *testing.T) {
		doc := "```\n\tcode\n```"
		if errs := CheckNoHardTabs("t.md", scanStyleDoc(doc), 0); len(errs) != 0 {
			t.Errorf("fenced: got %d, want 0: %+v", len(errs), errs)
		}
	})
}
