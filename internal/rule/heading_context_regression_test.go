package rule

import (
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// scanDoc is a small helper: it trims a leading newline (so table literals can
// start on their own line) and returns the scanned context.
func scanDoc(doc string) *preprocess.Context {
	doc = strings.TrimPrefix(doc, "\n")
	return preprocess.Scan(strings.Split(doc, "\n"))
}

// TestHeadingFamilySkipsBlockContexts is the #337 Phase 3 (heading family)
// gap-closure regression suite. Before migrating to the preprocess context these
// rules skipped only fenced code, so heading-, setext-, and emphasis-looking
// lines inside indented code blocks, HTML blocks, and HTML comments were treated
// as real content (false positives, and for heading-level state pollution). Each
// case feeds such lines and asserts no violation is produced.
func TestHeadingFamilySkipsBlockContexts(t *testing.T) {
	// Each doc places heading/emphasis/setext-looking lines inside a block
	// context. The only "real" heading (when one is needed to keep the rule's
	// happy path valid) lives outside the block.
	indentedCode := "\nReal intro\n\n    # Fake heading\n    ## Another fake\n\nmore text\n"
	htmlBlock := "\n<div>\n# Not a heading\n## Also not\n</div>\n"
	htmlComment := "\n<!--\n# commented heading\n## also commented\n-->\n"

	t.Run("single-h1", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "\n# Real H1\n\n    # Fake H1\n",
			"html block":   "\n# Real H1\n\n<div>\n# Fake H1\n</div>\n",
			"html comment": "\n# Real H1\n\n<!--\n# Fake H1\n-->\n",
		} {
			if errs := CheckSingleH1("t.md", scanDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("heading-level", func(t *testing.T) {
		// A level jump that lives inside a block must not be reported, and a
		// fake heading must not pollute the running level (minLevel 1).
		for name, doc := range map[string]string{
			"indented":     "\n# Intro\n\n    #### Fake jump\n\ntext\n",
			"html block":   "\n# Intro\n\n<div>\n#### Fake jump\n</div>\n",
			"html comment": "\n# Intro\n\n<!--\n#### Fake jump\n-->\n",
		} {
			if errs := CheckHeadingLevels("t.md", scanDoc(doc), 0, 1); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("duplicate-heading", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "\n# Title\n\n    # Title\n",
			"html block":   "\n# Title\n\n<div>\n# Title\n</div>\n",
			"html comment": "\n# Title\n\n<!--\n# Title\n-->\n",
		} {
			if errs := CheckDuplicateHeadings("t.md", scanDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("no-setext-headings", func(t *testing.T) {
		// An indented underline is already excluded by the setext regex (max 3
		// leading spaces), so the meaningful gaps are HTML block and comment,
		// where the underline sits at column 0.
		for name, doc := range map[string]string{
			"html block":   "\n<div>\nHeading\n===\n</div>\n",
			"html comment": "\n<!--\nHeading\n===\n-->\n",
		} {
			if errs := CheckNoSetextHeadings("t.md", scanDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("blanks-around-headings", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     indentedCode,
			"html block":   htmlBlock,
			"html comment": htmlComment,
		} {
			if errs := CheckBlanksAroundHeadings("t.md", scanDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})

	t.Run("no-emphasis-as-heading", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "\npara\n\n    *Emphasis*\n\ntext\n",
			"html block":   "\n<div>\n*Emphasis*\n</div>\n",
			"html comment": "\n<!--\n*Emphasis*\n-->\n",
		} {
			if errs := CheckNoEmphasisAsHeading("t.md", scanDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
	})
}
