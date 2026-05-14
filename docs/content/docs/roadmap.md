---
title: "Roadmap"
weight: 9
---

# Roadmap

## Rules — Completed

### Link checks

- [x] `external-link`: Validate external HTTP/HTTPS URLs
- [x] `link-fragments`: Internal anchor links must resolve to an existing heading

### Structure and formatting

All Priority 1 rules from the [ecosystem analysis (#76)](https://github.com/shinagawa-web/gomarklint/issues/76) have landed.

- [x] `no-multiple-blank-lines`: Disallow multiple consecutive blank lines
- [x] `fenced-code-language`: Fenced code blocks must specify a language
- [x] `single-h1`: Only one H1 heading per document
- [x] `blanks-around-headings`: Headings must be surrounded by blank lines
- [x] `no-bare-urls`: URLs must use proper link syntax, not bare text
- [x] `no-empty-links`: Links must not have an empty destination
- [x] `no-emphasis-as-heading`: Bold/italic must not substitute for headings
- [x] `no-setext-headings`: Setext-style headings must use ATX style instead
- [x] `blanks-around-lists`: Lists must be surrounded by blank lines
- [x] `max-line-length`: Enforce a maximum line length
- [x] `no-hard-tabs`: No hard tab characters in body text
- [x] `blanks-around-fences`: Fenced code blocks must be surrounded by blank lines
- [x] `consistent-code-fence`: Consistent fence character (`` ``` `` vs `~~~`)
- [x] `no-trailing-punctuation`: No trailing punctuation in headings
- [x] `consistent-emphasis-style`: Consistent emphasis marker (`*` vs `_`)
- [x] `consistent-list-marker`: Consistent unordered list marker (`-` vs `*` vs `+`)

## Rules — Planned

- [ ] `no-trailing-spaces`: No trailing whitespace at end of lines
- [ ] `no-undefined-references`: Reference-style links/images must have a matching definition
- [ ] `table-formatting`: Table structure and cell-padding consistency
- [ ] `descriptive-link-text`: Link text must not be generic ("click here", "here")
- [ ] `consistent-line-endings`: Enforce consistent line endings (LF vs CRLF)

## Extensibility

- [x] Allow disabling rules via inline comments (e.g. `<!-- gomarklint-disable -->`)
- [ ] Plugin system for custom rules (via Go interface or external binary)

## Distribution & CI

- [x] GitHub Actions integration
- [x] Prebuilt binaries via `goreleaser` (macOS/Linux/Windows)
- [x] Homebrew tap (`brew install shinagawa-web/tap/gomarklint`)
- [x] npm package (`npm install -g @shinagawa-web/gomarklint`)

## Developer UX

- [x] Rule severity levels (`error` / `warning` / `off`)
- [ ] Rule messages with IDs and documentation links
- [ ] File caching for faster repeated linting
- [ ] VS Code extension using gomarklint core
- [ ] Interactive mode (e.g. prompt to fix or explain errors)

## Internationalization

- [ ] Localized error messages (e.g. Japanese, Spanish)

---

Feel free to suggest more ideas by [opening an issue or discussion on GitHub](https://github.com/shinagawa-web/gomarklint/issues)!
