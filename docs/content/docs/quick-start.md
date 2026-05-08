---
title: "Quick Start"
weight: 1
---

# Quick Start

## Install

**Quick install** (macOS / Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | sh
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

**Download binary** (no Go required):

Download the latest binary for your platform from [GitHub Releases](https://github.com/shinagawa-web/gomarklint/releases/latest).

```sh
# macOS (Intel)
tar -xzf gomarklint_Darwin_x86_64.tar.gz
# Linux (x86_64)
tar -xzf gomarklint_Linux_x86_64.tar.gz
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

**Build from source:**

```sh
git clone https://github.com/shinagawa-web/gomarklint
cd gomarklint
make build   # or: go build -o gomarklint .
```

## 1) Initialize config (optional but recommended)

```sh
gomarklint init
```

This creates `.gomarklint.json` with sensible defaults:

```json
{
  "default": true,
  "rules": {
    "final-blank-line": true,
    "unclosed-code-block": true,
    "empty-alt-text": true,
    "fenced-code-language": true,
    "heading-level": { "severity": "error", "minLevel": 2 },
    "duplicate-heading": true,
    "no-multiple-blank-lines": true,
    "no-setext-headings": true,
    "single-h1": true,
    "blanks-around-headings": true,
    "no-bare-urls": true,
    "no-empty-links": true,
    "no-emphasis-as-heading": true,
    "blanks-around-lists": true,
    "blanks-around-fences": true,
    "no-hard-tabs": true,
    "no-trailing-punctuation": { "punctuation": ".,;:!" },
    "consistent-code-fence": { "style": "consistent" },
    "consistent-emphasis-style": { "style": "consistent" },
    "consistent-list-marker": { "style": "consistent" },
    "max-line-length": { "enabled": false, "lineLength": 80 },
    "external-link": { "enabled": false, "severity": "error", "timeoutSeconds": 5, "skipPatterns": [] },
    "link-fragments": { "enabled": true, "slug-algorithm": "github" }
  },
  "include": ["README.md", "testdata"],
  "ignore": [],
  "output": "text"
}
```

You can edit it anytime — CLI flags override config values.

## 2) Run it

```sh
# check current directory recursively
gomarklint .

# check specific targets
gomarklint docs README.md internal/rule
```

Exit code is non-zero if any violations are found, zero otherwise.

## 3) JSON output (for CI / tooling)

```sh
gomarklint . --output json
```
