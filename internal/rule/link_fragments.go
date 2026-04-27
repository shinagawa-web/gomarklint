package rule

import (
	"fmt"
	"regexp"
	"strings"
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
	if rest[0] != ' ' {
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

// collectRefDefs returns a map from normalized reference label to fragment string (without #)
// for all reference link definitions in lines that target a fragment destination.
func collectRefDefs(lines []string) map[string]string {
	defs := make(map[string]string)
	for _, line := range lines {
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

// collectHeadingSlugs returns the set of all valid fragment slugs computed from ATX headings
// in lines. Headings inside fenced code blocks are excluded.
// Duplicate headings produce suffixed slugs (-1, -2, …) following GitHub convention.
func collectHeadingSlugs(lines []string, algorithm string) map[string]struct{} {
	var headings []string

	inBlock := false
	fenceMarker := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}
		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		text, level := extractHeadingText(trimmed)
		if level == 0 {
			continue
		}
		headings = append(headings, text)
	}

	return buildSlugSet(headings, algorithm)
}

// CheckLinkFragments validates that every internal fragment link in the document
// resolves to an actual heading slug. Both inline links ([text](#frag)) and
// reference links ([text][ref] + [ref]: #frag) are checked. Content inside fenced
// code blocks and inline code spans is skipped.
//
// Supported options:
//   - "slug-algorithm": string — one of github, gitlab, zenn, pandoc, pandoc-gfm,
//     kramdown, mkdocs, docfx, hugo (default: "github")
func CheckLinkFragments(filename string, lines []string, offset int, options map[string]interface{}) []LintError {
	algorithm := "github"
	if v, ok := options["slug-algorithm"]; ok {
		if s, ok := v.(string); ok && s != "" {
			algorithm = s
		}
	}

	slugs := collectHeadingSlugs(lines, algorithm)
	refDefs := collectRefDefs(lines)

	var errs []LintError
	inBlock := false
	fenceMarker := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inBlock {
			if IsClosingFence(trimmed, fenceMarker) {
				inBlock = false
				fenceMarker = ""
			}
			continue
		}
		if marker := openingFenceMarker(trimmed); marker != "" {
			inBlock = true
			fenceMarker = marker
			continue
		}

		scanned := line
		if strings.ContainsRune(scanned, '`') {
			scanned = stripInlineCode(scanned)
		}

		lineNum := offset + i + 1

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
	}

	return errs
}
