package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

// extractHeadingText extracts the plain text from an ATX heading line (e.g. ## Heading).
// Returns the trimmed heading text and its level (1–6).
// Returns "", 0 for lines that are not valid ATX headings.
func extractHeadingText(line string) (string, int) {
	if len(line) == 0 || line[0] != '#' {
		return "", 0
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level > 6 {
		return "", 0
	}
	rest := line[level:]
	if rest == "" {
		return "", level
	}
	if rest[0] != ' ' && rest[0] != '\t' {
		return "", 0
	}
	return strings.TrimSpace(rest[1:]), level
}

// reFragmentLink matches inline fragment links: [text](#fragment)
// Capture group 1 is the fragment without the leading #.
var reFragmentLink = regexp.MustCompile(`\[[^\]]*\]\(#([^)]+)\)`)

// reRefLinkUsage matches reference-style link usage: [text][label]
// Capture group 1 is the reference label.
var reRefLinkUsage = regexp.MustCompile(`\[[^\]]*\]\[([^\]]+)\]`)

// reRefDef matches reference link definitions that target a fragment: [label]: #fragment
// Capture group 1 is the label; group 2 is the fragment without the leading #.
var reRefDef = regexp.MustCompile(`^\s*\[([^\]]+)\]:\s+#(\S+)`)

// reStripInlineImages matches inline images: ![alt](url)
// Used to remove image syntax before fragment link detection to avoid false positives
// from image fragments like ![alt](#fig-1).
var reStripInlineImages = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)

// collectRefDefs returns a map from normalized reference label to fragment string (without #)
// for all reference link definitions that target a fragment destination.
// Definitions inside fenced/indented code, HTML blocks, and HTML comments are excluded.
func collectRefDefs(ctx *preprocess.Context) map[string]string {
	defs := make(map[string]string)
	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Line(i)
		// Reference definitions must start with optional whitespace then '['.
		if !strings.HasPrefix(strings.TrimLeft(line, " \t"), "[") {
			continue
		}
		m := reRefDef.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(m[1]))
		fragment := strings.TrimSpace(m[2])
		defs[label] = fragment
	}
	return defs
}

// collectHeadingSlugs returns the set of all valid fragment slugs computed from ATX headings.
// Headings inside fenced/indented code, HTML blocks, and HTML comments are
// excluded, so a fake heading inside (e.g.) an indented code block no longer
// pollutes the valid-slug set and masks a broken fragment link (audit #337 false
// negative).
// Duplicate headings produce suffixed slugs (-1, -2, …) following GitHub convention.
func collectHeadingSlugs(ctx *preprocess.Context, slugger func(string) string) map[string]struct{} {
	var headings []string

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		text, level := extractHeadingText(strings.TrimSpace(ctx.Line(i)))
		if level == 0 {
			continue
		}
		headings = append(headings, text)
	}

	return buildSlugSet(headings, slugger)
}

// hasAnyFragmentSyntax is a cheap pre-filter that reports whether any non-block
// line could be a fragment link or fragment definition. Checks for "(#" (inline
// fragment links) and "]: #" (reference definition pointing at a fragment).
//
// It skips block contexts, since every real pass below does too — so fragment
// syntax that lives only inside code/HTML/comment no longer trips the more
// expensive passes. It reads the raw line (a superset of the sanitized text and
// of the raw text the ref-def pass uses), so it never exits early while a real
// fragment construct exists.
func hasAnyFragmentSyntax(ctx *preprocess.Context) bool {
	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Line(i)
		if strings.Contains(line, "(#") || strings.Contains(line, "]: #") {
			return true
		}
	}
	return false
}

// hasAnyFragmentLinks reports whether any non-block line contains at least one
// potential fragment link. Like hasAnyFragmentSyntax it skips block contexts and
// reads the raw line, staying O(n) with a single pass and minimal overhead.
func hasAnyFragmentLinks(ctx *preprocess.Context, refDefs map[string]string) bool {
	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Line(i)
		if strings.Contains(line, "(#") {
			return true
		}
		if len(refDefs) > 0 && strings.Contains(line, "][") {
			return true
		}
	}
	return false
}

// parseSlugAlgorithm extracts the slug-algorithm option, defaulting to "github".
func parseSlugAlgorithm(options map[string]interface{}) string {
	if v, ok := options["slug-algorithm"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "github"
}

// checkInlineFragments checks inline fragment links ([text](#frag)) on one line.
func checkInlineFragments(filename string, lineNum int, scanned string, slugs map[string]struct{}) []LintError {
	var errs []LintError
	for _, m := range reFragmentLink.FindAllStringSubmatch(scanned, -1) {
		fragment := m[1]
		if _, ok := slugs[fragment]; !ok {
			errs = append(errs, LintError{
				File:    filename,
				Line:    lineNum,
				Message: fmt.Sprintf("link-fragments: fragment #%s not found in this document", fragment),
			})
		}
	}
	return errs
}

// checkRefFragments checks reference-style fragment links ([text][ref] with [ref]: #frag) on one line.
func checkRefFragments(filename string, lineNum int, scanned string, slugs map[string]struct{}, refDefs map[string]string) []LintError {
	var errs []LintError
	for _, m := range reRefLinkUsage.FindAllStringSubmatch(scanned, -1) {
		label := strings.ToLower(strings.TrimSpace(m[1]))
		fragment, ok := refDefs[label]
		if !ok {
			continue
		}
		if _, found := slugs[fragment]; !found {
			errs = append(errs, LintError{
				File:    filename,
				Line:    lineNum,
				Message: fmt.Sprintf("link-fragments: fragment #%s not found in this document", fragment),
			})
		}
	}
	return errs
}

// CheckLinkFragments validates that every internal fragment link in the document
// resolves to an actual heading slug. Both inline links ([text](#frag)) and
// reference links ([text][ref] + [ref]: #frag) are checked. Content inside fenced
// code, indented code, HTML blocks, HTML comments, and inline code spans is
// skipped — for both the link checks and the heading/definition collection.
//
// Supported options:
//   - "slug-algorithm": string — preset name (github, gitlab, zenn, pandoc, pandoc-gfm,
//     kramdown, mkdocs, docfx, hugo, qiita, mdbook, vitepress, gitea, forgejo, sphinx,
//     eleventy, azure-devops, myst, docusaurus, gatsby, astro, starlight, nuxt-content,
//     quarto) or "custom" (default: "github")
//   - "slug-params": map — used only when slug-algorithm is "custom"; keys: lowercase (bool),
//     preserve-unicode (bool), space-replacement (string), strip-chars (regex string),
//     collapse-separators (bool)
func CheckLinkFragments(filename string, ctx *preprocess.Context, offset int, options map[string]interface{}) []LintError {
	if !hasAnyFragmentSyntax(ctx) {
		return nil
	}
	refDefs := collectRefDefs(ctx)
	if !hasAnyFragmentLinks(ctx, refDefs) {
		return nil
	}
	algorithm := parseSlugAlgorithm(options)
	slugs := collectHeadingSlugs(ctx, makeSlugger(algorithm, options))

	var errs []LintError

	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}

		// Sanitized blanks inline code spans (and inline comments); inline images
		// are stripped separately so image fragments like ![alt](#fig) are not
		// treated as broken links.
		scanned := ctx.Sanitized(i)
		if strings.ContainsRune(scanned, '!') {
			scanned = reStripInlineImages.ReplaceAllString(scanned, "")
		}

		hasInlineLink := strings.Contains(scanned, "(#")
		hasRefLink := len(refDefs) > 0 && strings.Contains(scanned, "][")
		if !hasInlineLink && !hasRefLink {
			continue
		}

		lineNum := offset + i + 1
		if hasInlineLink {
			errs = append(errs, checkInlineFragments(filename, lineNum, scanned, slugs)...)
		}
		if hasRefLink {
			errs = append(errs, checkRefFragments(filename, lineNum, scanned, slugs, refDefs)...)
		}
	}

	return errs
}
