package rule

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// reSlugHTMLComment matches HTML comments for removal from heading text.
var reSlugHTMLComment = regexp.MustCompile(`<!--.*?-->`)

// reSlugHTMLTag matches HTML tags for removal from heading text.
var reSlugHTMLTag = regexp.MustCompile(`<[^>]+>`)

// reSlugRefImage matches reference-style images: ![alt][ref]
var reSlugRefImage = regexp.MustCompile(`!\[([^\]]*)\]\[[^\]]*\]`)

// reSlugImage matches inline images: ![alt](url)
var reSlugImage = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)

// reSlugRefLink matches reference-style links: [text][ref]
var reSlugRefLink = regexp.MustCompile(`\[([^\]]*)\]\[[^\]]*\]`)

// reSlugLink matches inline links: [text](url)
var reSlugLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// reSlugBoldAsterisk matches **bold**
var reSlugBoldAsterisk = regexp.MustCompile(`\*\*([^*]+)\*\*`)

// reSlugBoldUnderscore matches __bold__
var reSlugBoldUnderscore = regexp.MustCompile(`__([^_]+)__`)

// reSlugItalicAsterisk matches *italic*
var reSlugItalicAsterisk = regexp.MustCompile(`\*([^*]+)\*`)

// reSlugItalicUnderscore matches _italic_
var reSlugItalicUnderscore = regexp.MustCompile(`_([^_]+)_`)

// reSlugCode matches inline code spans (single or multi-backtick), keeping content.
var reSlugCode = regexp.MustCompile("`+([^`]+)`+")

// reSphinxNonAlnum replaces runs of non-alphanumeric ASCII chars with a single hyphen.
var reSphinxNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// stripHeadingFormatting removes Markdown and HTML inline formatting from heading text,
// returning plain text for slug generation.
// Order: HTML comments → HTML tags → images → links → code spans (saved as placeholders) →
// bold → italic → code span content restored.
// Code spans are saved before bold/italic to prevent underscores/asterisks inside backticks
// from being mis-parsed as emphasis markers.
func stripHeadingFormatting(s string) string {
	// Fast path: plain headings with no formatting markers need no processing.
	if !strings.ContainsAny(s, "*_[<`!") {
		return s
	}
	s = reSlugHTMLComment.ReplaceAllString(s, "")
	s = reSlugHTMLTag.ReplaceAllString(s, "")
	s = reSlugRefImage.ReplaceAllString(s, "$1")
	s = reSlugImage.ReplaceAllString(s, "$1")
	s = reSlugRefLink.ReplaceAllString(s, "$1")
	s = reSlugLink.ReplaceAllString(s, "$1")
	// Save code span content before bold/italic so that underscores/asterisks
	// inside backtick spans are not consumed by the emphasis regexes.
	var saved []string
	s = reSlugCode.ReplaceAllStringFunc(s, func(m string) string {
		idx := len(saved)
		saved = append(saved, strings.Trim(m, "`"))
		return fmt.Sprintf("\x00%d\x00", idx)
	})
	s = reSlugBoldAsterisk.ReplaceAllString(s, "$1")
	s = reSlugBoldUnderscore.ReplaceAllString(s, "$1")
	s = reSlugItalicAsterisk.ReplaceAllString(s, "$1")
	s = reSlugItalicUnderscore.ReplaceAllString(s, "$1")
	for i, content := range saved {
		s = strings.ReplaceAll(s, fmt.Sprintf("\x00%d\x00", i), content)
	}
	return s
}

// githubStripRune reports whether r should be removed by the GitHub slug algorithm.
// Matches github-slugger v2: keeps \p{L}, \p{Nd}, \p{Nl}, hyphens, and underscores.
func githubStripRune(r rune) bool {
	if r == '-' || r == '_' {
		return false
	}
	return !unicode.IsLetter(r) && !unicode.Is(unicode.Nd, r) && !unicode.Is(unicode.Nl, r)
}

// slugGitHub computes the GitHub-compatible slug (github-slugger v2).
// Lowercases, strips all runes outside \p{L}/\p{Nd}/\p{Nl}/-/_, replaces whitespace with hyphens.
// Consecutive whitespace produces consecutive hyphens (no collapsing).
func slugGitHub(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if !githubStripRune(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// slugGitLab computes the GitLab (goldmark slugify) slug.
// Lowercases, keeps Unicode letters/numbers/hyphens/underscores, collapses consecutive hyphens.
func slugGitLab(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	return collapseDashes(sb.String())
}

// slugZenn computes the Zenn (markdown-it-anchor default) slug.
// Lowercases and replaces whitespace with hyphens; all other characters are preserved.
func slugZenn(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// slugPandoc computes the Pandoc auto_identifiers slug.
// Lowercases, keeps only ASCII letters/digits/hyphens/underscores/periods, collapses consecutive
// hyphens, then strips everything up to the first letter (Pandoc auto_identifiers step 5).
func slugPandoc(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			sb.WriteRune(r)
		}
	}
	s := collapseDashes(sb.String())
	for i, r := range s {
		if r >= 'a' && r <= 'z' {
			return s[i:]
		}
	}
	return ""
}

// slugKramdown computes the kramdown header_ids slug.
// Lowercases, keeps only ASCII letters/digits/hyphens, replaces spaces with hyphens, collapses,
// then strips leading digits and hyphens (kramdown skips chars until the first letter).
func slugKramdown(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-':
			sb.WriteRune(r)
		case unicode.IsSpace(r):
			sb.WriteByte('-')
		}
	}
	result := collapseDashes(sb.String())
	return strings.TrimLeft(result, "0123456789-")
}

// slugMkDocs computes the MkDocs (Python-Markdown toc.py) slug.
// Lowercases, strips non-ASCII characters, replaces spaces with hyphens, collapses.
func slugMkDocs(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if r < 128 && ((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			sb.WriteRune(r)
		}
	}
	return collapseDashes(sb.String())
}

// slugDocFX computes the DocFX (Markdig AutoIdentifiers) slug.
// Preserves case, keeps only [a-zA-Z0-9-_.], replaces spaces with hyphens, collapses.
func slugDocFX(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' {
			sb.WriteRune(r)
		}
	}
	return collapseDashes(sb.String())
}

// nfkdStripCombining applies NFKD normalization and removes Unicode combining characters (category Mn).
// This converts precomposed accented characters to their base ASCII equivalents
// (e.g. 'é' → 'e', 'ü' → 'u') while preserving CJK and other non-combining Unicode.
func nfkdStripCombining(text string) string {
	normalized := norm.NFKD.String(text)
	var sb strings.Builder
	sb.Grow(len(normalized))
	for _, r := range normalized {
		if !unicode.Is(unicode.Mn, r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// slugQiita computes the Qiita slug.
// Equivalent to Ruby: downcase.gsub(/[^\p{Word}\- ]/u, "").tr(" ", "-")
// Keeps Unicode letters, digits, underscores, and hyphens; collapses nothing.
func slugQiita(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// slugVitePress computes the VitePress (markdown-it-anchor) slug.
// NFKD-normalizes to strip combining chars (accented Latin → ASCII base),
// then lowercases and replaces non-alphanumeric chars with hyphens, collapsing runs.
// CJK and other Unicode letters/digits are preserved.
func slugVitePress(text string) string {
	text = nfkdStripCombining(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteByte('-')
		}
	}
	return collapseDashes(sb.String())
}

// slugGitea computes the Gitea (goldmark) slug.
// Same as GitHub algorithm but prefixed with "user-content-".
func slugGitea(text string) string {
	return "user-content-" + slugGitHub(text)
}

// slugSphinx computes the Sphinx (Python-Sphinx auto-section-label) slug.
// NFKD-normalizes to strip combining chars, keeps only lowercase ASCII alphanumerics,
// replaces runs of non-alphanumeric with a single hyphen, then strips leading hyphens and
// leading digits (matching docutils _non_id_at_ends: ^[-0-9]+|-+$) and trailing hyphens.
// Non-Latin-only headings that produce an empty result return "" (the id1/id2 fallback
// is document-level state that cannot be reproduced outside the build context).
func slugSphinx(text string) string {
	text = nfkdStripCombining(text)
	var ascii strings.Builder
	ascii.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if r < 128 {
			ascii.WriteRune(r)
		}
	}
	result := reSphinxNonAlnum.ReplaceAllString(ascii.String(), "-")
	// docutils _non_id_at_ends strips leading hyphens AND leading digits, trailing hyphens only.
	result = strings.TrimLeft(result, "-0123456789")
	return strings.TrimRight(result, "-")
}

// umlautReplacer expands German umlauts before NFKD, matching @sindresorhus/slugify's char map.
var umlautReplacer = strings.NewReplacer(
	"ä", "ae", "Ä", "ae",
	"ö", "oe", "Ö", "oe",
	"ü", "ue", "Ü", "ue",
	"ß", "ss",
)

// slugEleventy computes an approximation of the Eleventy (@sindresorhus/slugify) slug.
// Expands German umlauts, NFKD-normalizes, then converts any non-ASCII-alphanumeric char
// to a hyphen (collapsing consecutive ones). Non-ASCII chars without a mapping (CJK, etc.)
// also become hyphens, matching @sindresorhus/slugify which replaces unknown chars with the
// separator; leading/trailing hyphens are trimmed, so purely non-ASCII input yields "".
func slugEleventy(text string) string {
	text = umlautReplacer.Replace(text)
	text = nfkdStripCombining(text)
	var sb strings.Builder
	sb.Grow(len(text))
	prevWasHyphen := true // start true to suppress leading hyphens
	for _, r := range text {
		r = unicode.ToLower(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			prevWasHyphen = false
		} else if !prevWasHyphen {
			sb.WriteByte('-')
			prevWasHyphen = true
		}
	}
	return strings.TrimRight(sb.String(), "-")
}

// slugAzureDevOps computes the Azure DevOps Wiki slug.
// Lowercases, replaces Unicode space-separator chars (Zs) with hyphens, keeps RFC 3986
// unreserved chars (letters, digits, -, ., _, ~) as-is, and percent-encodes everything else.
func slugAzureDevOps(text string) string {
	const hexChars = "0123456789ABCDEF"
	var sb strings.Builder
	sb.Grow(len(text) * 2)
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.Is(unicode.Zs, r) {
			sb.WriteByte('-')
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '.' || r == '_' || r == '~' {
			sb.WriteRune(r)
		} else {
			// Percent-encode each UTF-8 byte (uppercase hex per RFC 3986).
			var buf [utf8.UTFMax]byte
			n := utf8.EncodeRune(buf[:], r)
			for _, b := range buf[:n] {
				sb.WriteByte('%')
				sb.WriteByte(hexChars[b>>4])
				sb.WriteByte(hexChars[b&0xf])
			}
		}
	}
	return sb.String()
}

// slugCustomParams holds the resolved parameters for the custom slug engine.
type slugCustomParams struct {
	lowercase          bool
	preserveUnicode    bool
	spaceReplacement   rune
	stripChars         *regexp.Regexp
	collapseSeparators bool
	collapseRe         *regexp.Regexp // pre-compiled collapse pattern; nil when collapse is off
}

// parseSlugParams extracts custom slug parameters from an options map.
func parseSlugParams(opts map[string]interface{}) slugCustomParams {
	p := slugCustomParams{
		lowercase:          true,
		preserveUnicode:    true,
		spaceReplacement:   '-',
		collapseSeparators: false,
	}
	params, _ := opts["slug-params"].(map[string]interface{})
	if params == nil {
		return p
	}
	if v, ok := params["lowercase"].(bool); ok {
		p.lowercase = v
	}
	if v, ok := params["preserve-unicode"].(bool); ok {
		p.preserveUnicode = v
	}
	if v, ok := params["space-replacement"].(string); ok && v != "" {
		runes := []rune(v)
		p.spaceReplacement = runes[0]
	}
	if v, ok := params["strip-chars"].(string); ok && v != "" {
		if re, err := regexp.Compile(v); err == nil {
			p.stripChars = re
		}
	}
	if v, ok := params["collapse-separators"].(bool); ok {
		p.collapseSeparators = v
	}
	if p.collapseSeparators {
		sep := string(p.spaceReplacement)
		if sep != "" {
			if re, err := regexp.Compile(regexp.QuoteMeta(sep) + "+"); err == nil {
				p.collapseRe = re
			}
		}
	}
	return p
}

// slugCustom applies the parameterized slug engine.
// Processing order: lowercase → per-char (space→sep, strip non-Unicode) → strip-chars regex → collapse.
func slugCustom(text string, p slugCustomParams) string {
	if p.lowercase {
		text = strings.ToLower(text)
	}
	sep := string(p.spaceReplacement)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			if sep != "" {
				sb.WriteRune(p.spaceReplacement)
			}
		} else if !p.preserveUnicode && r > 127 {
			// strip non-ASCII
		} else {
			sb.WriteRune(r)
		}
	}
	result := sb.String()
	if p.stripChars != nil {
		result = p.stripChars.ReplaceAllString(result, "")
	}
	if p.collapseRe != nil {
		result = p.collapseRe.ReplaceAllString(result, sep)
		result = strings.Trim(result, sep)
	}
	return result
}

// collapseDashes replaces runs of consecutive hyphens with a single hyphen
// and strips leading and trailing hyphens.
func collapseDashes(s string) string {
	if s == "" {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	prevWasDash := false
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			if !prevWasDash {
				sb.WriteByte('-')
			}
			prevWasDash = true
		} else {
			sb.WriteByte(s[i])
			prevWasDash = false
		}
	}
	result := sb.String()
	result = strings.TrimPrefix(result, "-")
	result = strings.TrimSuffix(result, "-")
	return result
}

// buildSlugSet builds the full set of valid fragment slugs from the given headings,
// applying deduplication. The first occurrence of a slug is bare (e.g. "intro");
// subsequent duplicates get a numeric suffix (e.g. "intro-1", "intro-2").
func buildSlugSet(headings []string, slugger func(string) string) map[string]struct{} {
	slugs := make(map[string]struct{})
	seen := make(map[string]int)
	for _, text := range headings {
		plain := stripHeadingFormatting(text)
		base := slugger(plain)
		if base == "" {
			continue
		}
		count := seen[base]
		seen[base]++
		var slug string
		if count == 0 {
			slug = base
		} else {
			slug = fmt.Sprintf("%s-%d", base, count)
		}
		slugs[slug] = struct{}{}
	}
	return slugs
}

// makeSlugger returns a slug function for the given algorithm and options.
// For "custom", options["slug-params"] is parsed into slugCustomParams.
// All other algorithm names are dispatched through ComputeSlug.
func makeSlugger(algorithm string, options map[string]interface{}) func(string) string {
	if algorithm == "custom" {
		params := parseSlugParams(options)
		return func(text string) string {
			return slugCustom(text, params)
		}
	}
	return func(text string) string {
		return ComputeSlug(text, algorithm)
	}
}

// slugRegistry maps each supported platform name to its slug function.
// Each platform is listed independently so configuration is self-evident and
// individual entries can be updated if an algorithm diverges from its current mapping.
var slugRegistry = map[string]func(string) string{
	// GitHub-compatible
	"github":       slugGitHub,
	"hugo":         slugGitHub,
	"pandoc-gfm":   slugGitHub,
	"myst":         slugGitHub,
	"docusaurus":   slugGitHub,
	"gatsby":       slugGitHub,
	"astro":        slugGitHub,
	"starlight":    slugGitHub,
	"nuxt-content": slugGitHub,
	// GitLab
	"gitlab": slugGitLab,
	// Zenn
	"zenn": slugZenn,
	// Pandoc family
	"pandoc": slugPandoc,
	"quarto": slugPandoc,
	// kramdown
	"kramdown": slugKramdown,
	// MkDocs
	"mkdocs": slugMkDocs,
	// DocFX
	"docfx": slugDocFX,
	// Qiita / mdBook (same character set)
	"qiita":  slugQiita,
	"mdbook": slugQiita,
	// VitePress
	"vitepress": slugVitePress,
	// Gitea / Forgejo
	"gitea":   slugGitea,
	"forgejo": slugGitea,
	// Sphinx
	"sphinx": slugSphinx,
	// Eleventy
	"eleventy": slugEleventy,
	// Azure DevOps
	"azure-devops": slugAzureDevOps,
}

// ComputeSlug generates a URL fragment slug from heading plain text using the named algorithm.
// Unknown algorithm names fall back to "github".
func ComputeSlug(text, algorithm string) string {
	if fn, ok := slugRegistry[algorithm]; ok {
		return fn(text)
	}
	return slugGitHub(text)
}
