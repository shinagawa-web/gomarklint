# gomarklint

[![npm](https://img.shields.io/npm/v/@shinagawa-web/gomarklint)](https://www.npmjs.com/package/@shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/shinagawa-web/gomarklint/blob/main/LICENSE)

> A fast, opinionated Markdown linter. Built in Go, designed for CI.

- Catch broken links and headings before your docs ship.
- Enforce predictable structure (no more "why is this H4 under H2?").
- Output that's friendly for both humans and machines (JSON).
- Process **100,000+ lines in ~170ms** — fast enough for local dev, robust enough for CI.

## Installation

```sh
npm install -g @shinagawa-web/gomarklint
```

The postinstall script automatically downloads the correct binary for your platform.

### Supported platforms

| OS      | x64 | arm64 |
|---------|-----|-------|
| macOS   | o   | o     |
| Linux   | o   | o     |
| Windows | o   | o     |

## Usage

```sh
# Lint all Markdown files in current directory
gomarklint

# Lint specific files
gomarklint README.md docs/**/*.md

# Output as JSON
gomarklint --format json
```

## Documentation

Full documentation is available at **[shinagawa-web.github.io/gomarklint](https://shinagawa-web.github.io/gomarklint/)**

- [Quick Start](https://shinagawa-web.github.io/gomarklint/docs/)
- [Rules](https://shinagawa-web.github.io/gomarklint/docs/rules/)
- [CLI Reference](https://shinagawa-web.github.io/gomarklint/docs/cli/)
- [Configuration](https://shinagawa-web.github.io/gomarklint/docs/configuration/)
- [GitHub Actions Integration](https://shinagawa-web.github.io/gomarklint/docs/github-actions/)

## Other installation methods

- **Shell:** `curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh`
- **Homebrew:** `brew install shinagawa-web/tap/gomarklint`
- **Go:** `go install github.com/shinagawa-web/gomarklint@latest`
- **Binary:** Download from [GitHub Releases](https://github.com/shinagawa-web/gomarklint/releases/latest)

## License

MIT License
