---
title: "gomarklint"
---

# gomarklint

[![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/shinagawa-web/gomarklint/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/shinagawa-web/gomarklint/badge)](https://securityscorecards.dev/viewer/?uri=github.com/shinagawa-web/gomarklint)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12970/badge)](https://www.bestpractices.dev/projects/12970)

<img src="demo.gif" width="800" alt="gomarklint catching a broken link and structure issues">

> Catch broken links before your readers do.
> A fast Markdown checker built in Go — validates internal anchors by default, external URLs when enabled, and enforces structure for local dev and CI alike.

```sh
# macOS / Linux — one-line install
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh

# or via Homebrew
brew install shinagawa-web/tap/gomarklint

# or via npm
npm install -g @shinagawa-web/gomarklint
```

[Quick Start →]({{< relref "/docs/quick-start" >}})

- Catch broken links before your readers do — validates internal anchors by default; enable `external-link` to also check external URLs.
- Enforce predictable structure (no more "why is this H4 under H2?").
- Output that's friendly for both humans and machines (JSON).

## Why

Docs break quietly and trust erodes loudly.
gomarklint focuses on reproducible rules that prevent "small but costly" failures:

- Heading hierarchies that drift during edits
- Duplicate headings that break anchor links
- Subtle dead links (including internal anchors)
- Large repos where "one-off checks" don't scale

> Goal: treat documentation quality like code quality—fast feedback locally, strict in CI, zero drama.

## Features

- **⚡️ Blazingly fast**: Process **100,000+ lines in ~170ms** (structural checks only, M4 Mac)
- Recursive `.md` search (multi-file & multi-directory)
- Frontmatter-aware parsing (YAML/TOML ignored when needed)
- File name & line number in diagnostics
- Human-readable and JSON outputs
- Fast single-binary CLI (Go), ideal for CI/CD
- Rules with clear rationales

Planned:

- Severity levels per rule
- Customizable rule enable/disable
- VS Code extension for in-editor feedback
