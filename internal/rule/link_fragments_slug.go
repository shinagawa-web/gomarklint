package rule

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
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

// stripHeadingFormatting removes Markdown and HTML inline formatting from heading text,
// returning plain text for slug generation.
// Order: HTML comments → HTML tags → images → links → bold → italic → code spans.
func stripHeadingFormatting(s string) string {
	s = reSlugHTMLComment.ReplaceAllString(s, "")
	s = reSlugHTMLTag.ReplaceAllString(s, "")
	s = reSlugRefImage.ReplaceAllString(s, "$1")
	s = reSlugImage.ReplaceAllString(s, "$1")
	s = reSlugRefLink.ReplaceAllString(s, "$1")
	s = reSlugLink.ReplaceAllString(s, "$1")
	s = reSlugBoldAsterisk.ReplaceAllString(s, "$1")
	s = reSlugBoldUnderscore.ReplaceAllString(s, "$1")
	s = reSlugItalicAsterisk.ReplaceAllString(s, "$1")
	s = reSlugItalicUnderscore.ReplaceAllString(s, "$1")
	s = reSlugCode.ReplaceAllString(s, "$1")
	return s
}

// githubStripRune reports whether r should be removed by the GitHub slug algorithm.
// Matches github-slugger v2: strips U+2000–U+206F, U+2E00–U+2E7F, and specific ASCII punctuation.
// Preserves hyphens, underscores, Unicode letters, and digits.
func githubStripRune(r rune) bool {
	if r >= 0x2000 && r <= 0x206F {
		return true // General Punctuation
	}
	if r >= 0x2E00 && r <= 0x2E7F {
		return true // Supplemental Punctuation
	}
	switch r {
	case '\\', '\'', '!', '"', '#', '$', '%', '&',
		'(', ')', '*', '+', ',', '.', '/', ':',
		';', '<', '=', '>', '?', '@', '[', ']',
		'^', '`', '{', '|', '}', '~':
		return true
	}
	return false
}

// slugGitHub computes the GitHub-compatible slug (github-slugger v2).
// Lowercases, strips specific punctuation ranges, replaces whitespace with hyphens.
// Consecutive whitespace produces consecutive hyphens (no collapsing).
func slugGitHub(text string) string {
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
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
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
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
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// slugPandoc computes the Pandoc auto_identifiers slug.
// Lowercases, keeps only ASCII letters/digits/hyphens/underscores, collapses consecutive hyphens.
func slugPandoc(text string) string {
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		if unicode.IsSpace(r) {
			sb.WriteByte('-')
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	return collapseDashes(sb.String())
}

// slugKramdown computes the kramdown header_ids slug.
// Lowercases, keeps only ASCII letters/digits/hyphens, replaces spaces with hyphens, collapses.
func slugKramdown(text string) string {
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-':
			sb.WriteRune(r)
		case unicode.IsSpace(r):
			sb.WriteByte('-')
		}
	}
	return collapseDashes(sb.String())
}

// slugMkDocs computes the MkDocs (Python-Markdown toc.py) slug.
// Lowercases, strips non-ASCII characters, replaces spaces with hyphens, collapses.
func slugMkDocs(text string) string {
	text = strings.ToLower(text)
	var sb strings.Builder
	sb.Grow(len(text))
	for _, r := range text {
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
func buildSlugSet(headings []string, algorithm string) map[string]struct{} {
	slugs := make(map[string]struct{})
	seen := make(map[string]int)
	for _, text := range headings {
		plain := stripHeadingFormatting(text)
		base := ComputeSlug(plain, algorithm)
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

// ComputeSlug generates a URL fragment slug from heading plain text using the named algorithm.
// Supported algorithms: github, gitlab, zenn, pandoc, pandoc-gfm, kramdown, mkdocs, docfx, hugo.
// Unknown algorithm names fall back to "github".
func ComputeSlug(text, algorithm string) string {
	switch algorithm {
	case "github", "hugo", "pandoc-gfm":
		return slugGitHub(text)
	case "gitlab":
		return slugGitLab(text)
	case "zenn":
		return slugZenn(text)
	case "pandoc":
		return slugPandoc(text)
	case "kramdown":
		return slugKramdown(text)
	case "mkdocs":
		return slugMkDocs(text)
	case "docfx":
		return slugDocFX(text)
	default:
		return slugGitHub(text)
	}
}
