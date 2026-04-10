---
title: "Quick Start"
weight: 1
---

# Quick Start

## Install

**Quick install** (macOS / Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/shinagawa-web/gomarklint/main/install.sh | bash
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
go install github.com/shinagawa-web/gomarklint@latest
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
  "minHeadingLevel": 2,
  "enableLinkCheck": false,
  "linkCheckTimeoutSeconds": 5,
  "skipLinkPatterns": [],
  "include": ["README.md", "testdata"],
  "ignore": [],
  "output": "text",
  "enableDuplicateHeadingCheck": true,
  "enableHeadingLevelCheck": true,
  "enableNoMultipleBlankLinesCheck": true,
  "enableNoSetextHeadingsCheck": true,
  "enableFinalBlankLineCheck": true
}
```

You can edit it anytime — CLI flags override config values.

## 2) Run it

```sh
# lint current directory recursively
gomarklint .

# lint specific targets
gomarklint docs README.md internal/rule
```

Exit code is non-zero if any violations are found, zero otherwise.

## 3) JSON output (for CI / tooling)

```sh
gomarklint . --output json
```
