# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

English | [日本語](README.ja.md)

<a href="https://gyazo.com/a5f8265a0865e5a37dc83733ca61069a"><img src="https://i.gyazo.com/a5f8265a0865e5a37dc83733ca61069a.gif" width="800" alt="Demo"></a>

> Blazing-fast Markdown linter built in Go — **100,000+ lines in ~170ms**. Single binary, no Node.js required, and built-in HTTP link validation.

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
go install github.com/shinagawa-web/gomarklint/v2@latest
```

- **100,000+ lines in ~170ms** — single binary, no JIT warmup, no runtime overhead.
- Catch broken links and headings before your docs ship.
- Enforce predictable structure (no more "why is this H4 under H2?").
- Output that's friendly for both humans and machines (JSON).

## CI Integration

### GitHub Actions

```yaml
name: gomarklint

on:
  push:
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: shinagawa-web/gomarklint-action@v1
        with:
          args: '.'
```

Full options and examples: [gomarklint-action](https://github.com/shinagawa-web/gomarklint-action)

### pre-commit

Add to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/shinagawa-web/gomarklint
    rev: v2.8.0
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
