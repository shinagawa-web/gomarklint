---
title: "Rules"
weight: 2
---

# Rules

`gomarklint` currently runs the following checks (ordered as executed):

| Rule key                       | What it detects                                                         | Notes / Options                                                                                       |
| ------------------------------ | ----------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `final-blank-line`             | Missing final blank line at EOF                                         | Default **on**                                                                                        |
| `unclosed-code-block`          | Unclosed fenced code blocks (`` ``` ``)                                 | Default **on**                                                                                        |
| `empty-alt-text`               | Image syntax with an empty alt text                                     | Default **on**                                                                                        |
| `heading-level`                | Invalid heading level progression (e.g., H2 → H4 skip)                 | Default **on**. Option: `minLevel` (default `2`)                                                      |
| `fenced-code-language`         | Fenced code blocks without a language identifier                        | Default **on**                                                                                        |
| `duplicate-heading`            | Duplicate headings within one file                                      | Default **on**                                                                                        |
| `no-multiple-blank-lines`      | Multiple consecutive blank lines                                        | Default **on**                                                                                        |
| `no-setext-headings`           | Setext heading used instead of ATX style                                | Default **on**                                                                                        |
| `single-h1`                    | More than one H1 heading in a file                                      | Default **on**                                                                                        |
| `blanks-around-headings`       | Headings not surrounded by blank lines                                  | Default **on**                                                                                        |
| `no-bare-urls`                 | HTTP/HTTPS URLs written as bare text instead of proper Markdown links   | Default **on**                                                                                        |
| `no-empty-links`               | Links or images with an empty destination (`[]()`, `[](#)`, `[](<>)`)  | Default **on**                                                                                        |
| `no-emphasis-as-heading`       | Bold/italic text used as a heading substitute instead of ATX headings   | Default **on**. Punctuation-ending spans (`. , ; : ! ? 。 、 ； ： ！ ？`) are excluded              |
| `blanks-around-lists`          | Lists not surrounded by blank lines                                     | Default **on**                                                                                        |
| `blanks-around-fences`         | Fenced code blocks not surrounded by blank lines                        | Default **on**                                                                                        |
| `no-hard-tabs`                 | Hard tab characters (`\t`) outside fenced code blocks and inline code   | Default **on**                                                                                        |
| `no-trailing-punctuation`      | Heading text ending with a punctuation character                        | Default **on**. Option: `punctuation` (default `".,;:!"`) — the full set of characters to flag; e.g. set `".,;:!?"` to also flag question headings |
| `max-line-length`              | Lines exceeding the configured maximum length                           | Default **off**. Option: `lineLength` (default `80`)                                                  |
| `external-link`                | External links that fail HTTP validation                                | Default **off**. Options: `timeoutSeconds` (default `5`), `skipPatterns` (regex list)                 |
| `link-fragments`               | Internal fragment links (`#section`) that do not resolve to a heading  | Default **off**. Options: `slug-algorithm` (default `github`), `slug-params` (for `custom` algorithm) |

## link-fragments

`link-fragments` validates that every internal fragment link in a document resolves to an actual heading slug. It supports multiple slug algorithms to match the platform where the Markdown is published.

### slug-algorithm

Set `slug-algorithm` to the name of the platform you are writing for. Each platform is an independent named value — you do not need to know which underlying algorithm it maps to.

**Supported platforms:**

| Value | Platform |
| --- | --- |
| `github` | GitHub (default) |
| `gitlab` | GitLab |
| `zenn` | Zenn |
| `qiita` | Qiita |
| `hugo` | Hugo |
| `vitepress` | VitePress |
| `docusaurus` | Docusaurus |
| `gatsby` | Gatsby (`gatsby-remark-autolink-headers`) |
| `astro` | Astro / Starlight |
| `starlight` | Starlight (alias for `astro`) |
| `nuxt-content` | Nuxt Content |
| `pandoc` | Pandoc (`auto_identifiers`) |
| `pandoc-gfm` | Pandoc (`gfm_auto_identifiers`) |
| `quarto` | Quarto |
| `kramdown` | kramdown |
| `mkdocs` | MkDocs |
| `docfx` | DocFX |
| `mdbook` | mdBook |
| `gitea` | Gitea |
| `forgejo` | Forgejo |
| `sphinx` | Sphinx |
| `eleventy` | Eleventy (`@sindresorhus/slugify`, approximate) |
| `azure-devops` | Azure DevOps Wiki |
| `myst` | MyST Parser |
| `custom` | Parameterized engine — see below |

### Custom algorithm

Use `slug-algorithm: "custom"` with `slug-params` for platforms not covered by the built-in presets:

```json
"link-fragments": {
  "enabled": true,
  "slug-algorithm": "custom",
  "slug-params": {
    "lowercase": true,
    "preserve-unicode": true,
    "space-replacement": "-",
    "strip-chars": "[^\\w\\- ]",
    "collapse-separators": true
  }
}
```

| Parameter | Type | Description |
| --- | --- | --- |
| `lowercase` | bool | Lowercase the heading before processing (default `true`) |
| `preserve-unicode` | bool | Keep non-ASCII characters in the slug (default `true`) |
| `space-replacement` | string | Character to replace spaces — `"-"` or `"_"` (default `"-"`) |
| `strip-chars` | string | Regex matching characters to remove after space replacement |
| `collapse-separators` | bool | Collapse consecutive separators and trim leading/trailing (default `false`) |

> **Note:** `strip-chars` uses Go's `regexp` syntax. `\w` matches ASCII `[0-9A-Za-z_]` only. To match Unicode word characters use `\p{L}`, `\p{N}`, etc.

## Execution details

- Files/dirs are expanded with ignore patterns from config.
- Per-file issues are sorted by line asc before printing.
- Line count is computed as `\n` count + 1 for reporting.
