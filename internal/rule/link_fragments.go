package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shinagawa-web/gomarklint/v3/internal/preprocess"
)

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

var reFragmentLink = regexp.MustCompile(`\[[^\]]*\]\(#([^)]+)\)`)
var reRefLinkUsage = regexp.MustCompile(`\[[^\]]*\]\[([^\]]+)\]`)
var reRefDef = regexp.MustCompile(`^\s*\[([^\]]+)\]:\s+#(\S+)`)
var reStripInlineImages = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)

func collectRefDefs(ctx *preprocess.Context) map[string]string {
	defs := make(map[string]string)
	for i := 0; i < ctx.Len(); i++ {
		if inBlockContext(ctx, i) {
			continue
		}
		line := ctx.Line(i)
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

func parseSlugAlgorithm(options map[string]interface{}) string {
	if v, ok := options["slug-algorithm"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "github"
}

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
