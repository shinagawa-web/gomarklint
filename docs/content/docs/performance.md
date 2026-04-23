---
title: "Performance"
weight: 7
---

# Performance

`gomarklint` is built for speed, with optimizations for both file parsing and external link validation.

## Structural checks

Headings, code blocks, blank lines, etc.:

- Scanning **185 files and 104,000+ lines** takes under **60ms**

## External link checking (`--enable-link-check`)

- Optimized concurrent validation with intelligent batching
- **~2,000 external links** validated in **under 10 seconds**
- Significantly faster than traditional sequential HTTP checks

## Recommended usage

**For rapid local feedback:**

- Run without `--enable-link-check` → completes in milliseconds
- Perfect for catching structural issues while editing

**For comprehensive validation:**

- Enable `--enable-link-check` for nightly CI runs, pre-release validation, or verifying newly added content

> **TL;DR:** Fast enough for local dev (no link check), robust enough for CI (with link check).
