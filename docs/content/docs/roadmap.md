---
title: "Roadmap"
weight: 8
---

# Roadmap (Post v1.0.0)

## Core Quality & Rule Expansion

- [ ] `max-line-length`: Enforce maximum line width
- [x] `no-multiple-consecutive-blank-lines`: Disallow multiple blank lines
- [ ] `image-alt-text` improvements: Enforce alt text style and length
- [ ] Rule severity levels (e.g. `warning`, `error`)

## Extensibility

- [ ] Plugin system for custom rules (via Go interface or external binary)
- [ ] Allow disabling specific rules via inline comments (e.g. `<!-- gomarklint-disable -->`)

## Testing & Stability

- [ ] Snapshot testing support for easier rule verification
- [ ] Regression test suite for real-world Markdown samples

## Developer UX

- [ ] VS Code extension using gomarklint core
- [ ] Interactive mode (e.g. prompt to fix or explain errors)
- [ ] File caching for faster repeated linting

## Ecosystem & CI

- [x] GitHub Actions integration
- [x] Prebuilt binaries via `goreleaser` (macOS/Linux/Windows)
- [ ] Homebrew formula
- [ ] Docker image (e.g. `ghcr.io/shinagawa-web/gomarklint`)

## Internationalization

- [ ] Localized messages (e.g. Japanese, Spanish)
- [ ] Rule messages with IDs and documentation links

---

Feel free to suggest more ideas by [opening an issue or discussion on GitHub](https://github.com/shinagawa-web/gomarklint/issues)!
