package rule

import (
	"strings"
	"testing"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

func scanLinkDoc(doc string) *preprocess.Context {
	doc = strings.TrimPrefix(doc, "\n")
	return preprocess.Scan(strings.Split(doc, "\n"))
}

// TestLinkFamilySkipsBlockContexts is the #337 Phase 3 (link/image family)
// gap-closure regression suite.
func TestLinkFamilySkipsBlockContexts(t *testing.T) {
	t.Run("empty-alt-text", func(t *testing.T) {
		// The audit worst offender: it previously fired in every context.
		for name, doc := range map[string]string{
			"fenced":       "```\n![](x.png)\n```",
			"indented":     "text\n\n    ![](x.png)",
			"inline code":  "see `![](x.png)` here",
			"html comment": "<!-- ![](x.png) -->",
			"html block":   "<div>\n![](x.png)\n</div>",
		} {
			if errs := CheckEmptyAltText("t.md", scanLinkDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
		// Sanity: a real empty-alt image in prose is still flagged.
		if errs := CheckEmptyAltText("t.md", scanLinkDoc("![](x.png)"), 0); len(errs) != 1 {
			t.Errorf("plain image: got %d errors, want 1", len(errs))
		}
	})

	t.Run("no-empty-links", func(t *testing.T) {
		for name, doc := range map[string]string{
			"indented":     "text\n\n    [a]()",
			"html comment": "<!-- [a]() -->",
			"html block":   "<div>\n[a]()\n</div>",
		} {
			if errs := CheckNoEmptyLinks("t.md", scanLinkDoc(doc), 0); len(errs) != 0 {
				t.Errorf("%s: got %d errors, want 0: %+v", name, len(errs), errs)
			}
		}
		if errs := CheckNoEmptyLinks("t.md", scanLinkDoc("[a]()"), 0); len(errs) != 1 {
			t.Errorf("plain empty link: got %d errors, want 1", len(errs))
		}
		// A non-empty link followed by trailing text: findEmptyLinks scans past
		// the link and finds no further "](", with no violation.
		if errs := CheckNoEmptyLinks("t.md", scanLinkDoc("[ok](http://x.com) and more text"), 0); len(errs) != 0 {
			t.Errorf("valid link with trailing text: got %d errors, want 0: %+v", len(errs), errs)
		}
	})

	t.Run("link-fragments false positive: link in indented code is not checked", func(t *testing.T) {
		// The fragment link lives in an indented code block, so it must not be
		// reported even though #nope resolves to nothing.
		doc := "# Real\n\n    [x](#nope)"
		if errs := CheckLinkFragments("t.md", scanLinkDoc(doc), 0, nil); len(errs) != 0 {
			t.Errorf("got %d errors, want 0: %+v", len(errs), errs)
		}
	})

	t.Run("link-fragments false negative: heading in indented code does not satisfy a link", func(t *testing.T) {
		// The only "# Fake" heading is inside an indented code block, so it must
		// NOT create a valid slug — the link to #fake is therefore broken.
		doc := "[link](#fake)\n\n    # Fake"
		errs := CheckLinkFragments("t.md", scanLinkDoc(doc), 0, nil)
		if len(errs) != 1 {
			t.Fatalf("got %d errors, want 1 (broken #fake link): %+v", len(errs), errs)
		}
	})

	t.Run("external-link extraction skips code/HTML/comment contexts", func(t *testing.T) {
		// ExtractExternalLinksWithLineNumbers feeds the network checker; URLs in
		// these contexts must not be extracted (and so never fetched).
		for name, doc := range map[string]string{
			"fenced":       "```\nhttp://x.com\n```",
			"indented":     "text\n\n    http://x.com",
			"inline code":  "see `http://x.com` here",
			"html comment": "<!-- http://x.com -->",
			"html block":   "<div>\nhttp://x.com\n</div>",
		} {
			got := ExtractExternalLinksWithLineNumbers(scanLinkDoc(doc), 0)
			if len(got) != 0 {
				t.Errorf("%s: extracted %d links, want 0: %+v", name, len(got), got)
			}
		}
		// Sanity: a bare URL in prose is still extracted.
		if got := ExtractExternalLinksWithLineNumbers(scanLinkDoc("see http://x.com today"), 0); len(got) != 1 {
			t.Errorf("plain URL: extracted %d, want 1", len(got))
		}
	})
}
