package rule

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var reSlugHTMLComment = regexp.MustCompile(`<!--.*?-->`)
var reSlugHTMLTag = regexp.MustCompile(`<[^>]+>`)
var reSlugRefImage = regexp.MustCompile(`!\[([^\]]*)\]\[[^\]]*\]`)
var reSlugImage = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
var reSlugRefLink = regexp.MustCompile(`\[([^\]]*)\]\[[^\]]*\]`)
var reSlugLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
var reSlugBoldAsterisk = regexp.MustCompile(`\*\*([^*]+)\*\*`)
var reSlugBoldUnderscore = regexp.MustCompile(`__([^_]+)__`)
var reSlugItalicAsterisk = regexp.MustCompile(`\*([^*]+)\*`)
var reSlugItalicUnderscore = regexp.MustCompile(`_([^_]+)_`)
var reSlugCode = regexp.MustCompile("`+([^`]+)`+")
var reSphinxNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

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

func githubStripRune(r rune) bool {
	if r == '-' || r == '_' {
		return false
	}
	return !unicode.IsLetter(r) && !unicode.Is(unicode.Nd, r) && !unicode.Is(unicode.Nl, r)
}

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

func slugMkDocs(text string) string {
	text = nfkdStripCombining(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	return collapseDashes(sb.String())
}

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

func slugMdBook(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))
	prevWasHyphen := true // start true to suppress any leading hyphen
	for _, r := range text {
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			sb.WriteRune(r)
			prevWasHyphen = false
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			if !prevWasHyphen {
				sb.WriteByte('-')
				prevWasHyphen = true
			}
		}
		// else: strip
	}
	return strings.TrimRight(sb.String(), "-")
}

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

// Gitea adds "user-content-" to DOM id for CSP isolation, but fragment links omit that prefix.
func slugGitea(text string) string {
	return slugGitHub(text)
}

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

var umlautReplacer = strings.NewReplacer(
	"ä", "ae", "Ä", "ae",
	"ö", "oe", "Ö", "oe",
	"ü", "ue", "Ü", "ue",
	"ß", "ss",
)

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

type slugCustomParams struct {
	lowercase          bool
	preserveUnicode    bool
	spaceReplacement   rune
	stripChars         *regexp.Regexp
	collapseSeparators bool
	collapseRe         *regexp.Regexp // pre-compiled collapse pattern; nil when collapse is off
}

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

var slugRegistry = map[string]func(string) string{
	"github":       slugGitHub,
	"hugo":         slugGitHub,
	"pandoc-gfm":   slugGitHub,
	"myst":         slugGitHub,
	"docusaurus":   slugGitHub,
	"gatsby":       slugGitHub,
	"astro":        slugGitHub,
	"starlight":    slugGitHub,
	"nuxt-content": slugGitHub,
	"gitlab":       slugGitLab,
	"zenn":         slugZenn,
	"pandoc":       slugPandoc,
	"quarto":       slugPandoc,
	"kramdown":     slugKramdown,
	"mkdocs":       slugMkDocs,
	"docfx":        slugDocFX,
	"qiita":        slugQiita,
	"mdbook":       slugMdBook,
	"vitepress":    slugVitePress,
	"gitea":        slugGitea,
	"forgejo":      slugGitea,
	"sphinx":       slugSphinx,
	"eleventy":     slugEleventy,
	"azure-devops": slugAzureDevOps,
}

func ComputeSlug(text, algorithm string) string {
	if fn, ok := slugRegistry[algorithm]; ok {
		return fn(text)
	}
	return slugGitHub(text)
}
