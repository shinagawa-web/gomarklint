package rule

import (
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func scanFenceDoc(doc string) *preprocess.Context {
	doc = strings.TrimPrefix(doc, "\n")
	return preprocess.Scan(strings.Split(doc, "\n"))
}

// TestFenceFamilySkipsBlockContexts is the #337 Phase 3 (fence family)
// gap-closure regression suite. A fence marker inside an indented code block or
// an HTML block is not a real fence, so none of the fence rules should treat it
// as one. Previously each rule did its own fence detection and mispaired these
// markers (audit #337, including the unclosed-code-block cascade).
func TestFenceFamilySkipsBlockContexts(t *testing.T) {
	// A lone ``` inside indented code (would previously open a phantom fence).
	indentedLoneFence := "\ntext\n\n    ```\n    code\n"
	// A no-language fence wholly inside an HTML block.
	htmlBlockFence := "\n<div>\n```\nno lang\n```\n</div>\n"
	// A tilde fence inside indented code (wrong marker, but not a real fence).
	indentedTildeFence := "\ntext\n\n    ~~~\n    code\n    ~~~\n"

	t.Run("unclosed-code-block ignores indented fence marker", func(t *testing.T) {
		if errs := CheckUnclosedCodeBlocks("t.md", scanFenceDoc(indentedLoneFence), 0); len(errs) != 0 {
			t.Errorf("got %d errors, want 0 (no phantom unclosed block): %+v", len(errs), errs)
		}
	})

	t.Run("fenced-code-language ignores fence in html block", func(t *testing.T) {
		if errs := CheckFencedCodeLanguage("t.md", scanFenceDoc(htmlBlockFence), 0); len(errs) != 0 {
			t.Errorf("got %d errors, want 0: %+v", len(errs), errs)
		}
	})

	t.Run("consistent-code-fence ignores fence in indented code", func(t *testing.T) {
		if errs := CheckConsistentCodeFence("t.md", scanFenceDoc(indentedTildeFence), 0, "backtick"); len(errs) != 0 {
			t.Errorf("got %d errors, want 0: %+v", len(errs), errs)
		}
	})

	t.Run("blanks-around-fences ignores fence in html block", func(t *testing.T) {
		if errs := CheckBlanksAroundFences("t.md", scanFenceDoc(htmlBlockFence), 0); len(errs) != 0 {
			t.Errorf("got %d errors, want 0: %+v", len(errs), errs)
		}
	})
}

// TestFenceFamilyAdjacentBlocks guards the adjacency case that a per-line
// InFencedCode flag cannot represent: two fenced blocks with no blank line
// between them. The closing line of the first and the opening line of the second
// are both inside fenced code, so a flag-run derivation would merge them into one
// block and miss the second opener / the missing blank between them.
func TestFenceFamilyAdjacentBlocks(t *testing.T) {
	t.Run("consistent-code-fence flags the second adjacent opener", func(t *testing.T) {
		// ``` block immediately followed by a ~~~ block: in "consistent" mode the
		// second opener's marker must be flagged.
		doc := "```\na\n```\n~~~\nb\n~~~"
		errs := CheckConsistentCodeFence("t.md", preprocess.Scan(strings.Split(doc, "\n")), 0, "consistent")
		if len(errs) != 1 || errs[0].Line != 4 {
			t.Fatalf("got %+v, want exactly 1 error on line 4 (the ~~~ opener)", errs)
		}
	})

	t.Run("blanks-around-fences flags the missing blank between adjacent fences", func(t *testing.T) {
		// Two backtick fences with nothing between them: the first needs a
		// trailing blank (line 3) and the second a leading blank (line 4).
		doc := "```\na\n```\n```\nb\n```"
		errs := CheckBlanksAroundFences("t.md", preprocess.Scan(strings.Split(doc, "\n")), 0)
		if len(errs) != 2 {
			t.Fatalf("got %d errors, want 2 (trailing for block 1, leading for block 2): %+v", len(errs), errs)
		}
	})

	t.Run("fenced-code-language checks both adjacent openers", func(t *testing.T) {
		// First fence has a language, second does not: only the second is flagged.
		doc := "```go\na\n```\n```\nb\n```"
		errs := CheckFencedCodeLanguage("t.md", preprocess.Scan(strings.Split(doc, "\n")), 0)
		if len(errs) != 1 || errs[0].Line != 4 {
			t.Fatalf("got %+v, want exactly 1 error on line 4 (second opener, no language)", errs)
		}
	})
}
