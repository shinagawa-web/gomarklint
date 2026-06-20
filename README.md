# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/shinagawa-web/gomarklint/badge)](https://securityscorecards.dev/viewer/?uri=github.com/shinagawa-web/gomarklint)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12970/badge)](https://www.bestpractices.dev/projects/12970)
[![npm version](https://img.shields.io/npm/v/@shinagawa-web/gomarklint.svg)](https://www.npmjs.com/package/@shinagawa-web/gomarklint)
[![npm downloads](https://img.shields.io/npm/dw/@shinagawa-web/gomarklint.svg)](https://www.npmjs.com/package/@shinagawa-web/gomarklint)

English | [日本語](README.ja.md)

<img src="docs/static/demo.gif" width="800" alt="gomarklint catching a broken link and structure issues">

> Catch broken links before your readers do — and keep your Markdown clean while you're at it. **100,000+ lines in ~170ms**, single binary, no Node.js required.

**Quick install** (macOS / Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh
```

**Download binary** (no Go required):

Download the latest binary for your platform from [GitHub Releases](https://github.com/shinagawa-web/gomarklint/releases/latest).

```sh
# macOS / Linux
tar -xzf gomarklint_Darwin_x86_64.tar.gz
sudo mv gomarklint /usr/local/bin/
# or install to user-local directory (no sudo required)
mkdir -p ~/.local/bin && mv gomarklint ~/.local/bin/
```

```powershell
# Windows (PowerShell)
Expand-Archive -Path gomarklint_Windows_x86_64.zip -DestinationPath "$env:LOCALAPPDATA\Programs\gomarklint"
# Add to PATH (run once)
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:LOCALAPPDATA\Programs\gomarklint", "User")
```

**Via Homebrew:**

```sh
brew install shinagawa-web/tap/gomarklint
```

**Via npm:**

```sh
npm install -g @shinagawa-web/gomarklint
```

**Via `go install`:**

```sh
go install github.com/shinagawa-web/gomarklint/v3@latest
```

- Catch broken links before your readers do — validates internal anchors by default; enable `external-link` to also check external URLs.
- **100,000+ lines in ~170ms** — single binary, no JIT warmup, no runtime overhead.
- Enforce predictable structure (no more "why is this H4 under H2?").
- Output that's friendly for both humans and machines (JSON).

## Concepts

gomarklint is built around four core concepts:

| Term | Description |
|---|---|
| **Rule** | A named, versioned check applied to a Markdown file, identified by a stable key (e.g. `heading-level`) |
| **Diagnostic** | A single violation emitted by a Rule, carrying file path, line number, rule key, message, and severity |
| **Severity** | The impact level of a violation: `error` causes a non-zero exit code; `warning` is informational only |
| **Config** | The per-rule and global settings that control enablement, severity, and rule-specific options |

Rule keys are stable identifiers — they will not be renamed without a major version bump.
Severity controls exit code behavior: any `error`-level Diagnostic makes gomarklint exit non-zero, making it safe as a CI gate.

See [docs/glossary.md](docs/glossary.md) for full definitions and Go type references.

## CI Integration

### GitHub Actions

[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-gomarklint-blue?logo=github)](https://github.com/marketplace/actions/gomarklint-markdown-linter)

```yaml
name: gomarklint

on:
  pull_request:
    paths:
      - '**/*.md'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: shinagawa-web/gomarklint@v3
```

Without `args`, the action lints the files listed in the `include` field of `.gomarklint.json`. Pass `args` to override, e.g. `args: docs/`.

Full options including PR comments: see [Marketplace listing](https://github.com/marketplace/actions/gomarklint-markdown-linter)

### pre-commit

Add to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/shinagawa-web/gomarklint
    rev: v3.0.0
    hooks:
      - id: gomarklint
```

## Documentation

Full documentation is available at **[shinagawa-web.github.io/gomarklint](https://shinagawa-web.github.io/gomarklint/)**

- [Quick Start](https://shinagawa-web.github.io/gomarklint/docs/quick-start/)
- [Rules](https://shinagawa-web.github.io/gomarklint/docs/rules/)
- [CLI Reference](https://shinagawa-web.github.io/gomarklint/docs/cli/)
- [Configuration](https://shinagawa-web.github.io/gomarklint/docs/configuration/)
- [GitHub Actions Integration](https://shinagawa-web.github.io/gomarklint/docs/github-actions/)
- [Migrating from Other Linters](https://shinagawa-web.github.io/gomarklint/docs/migration/)
- [FAQ & Troubleshooting](https://shinagawa-web.github.io/gomarklint/docs/faq/)

## Contributing

Issues, suggestions, and PRs are welcome!

Requirements: Go `1.22+` (latest stable recommended)

```sh
make test      # unit tests
make test-e2e  # end-to-end tests
make build     # build binary
```

### Git hooks

Install the pre-push hook to run lint and unit tests automatically before pushing:

```sh
make install-hooks
```

To bypass the hook in an emergency:

```sh
git push --no-verify
```

## License

MIT License
