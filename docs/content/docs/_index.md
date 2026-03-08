---
title: "Quick Start"
weight: 1
---

# Quick Start

## Install

**Via `go install`:**

```sh
go install github.com/shinagawa-web/gomarklint@latest
```

**Download prebuilt binary** (no Go required):

Download the latest binary for your platform from [GitHub Releases](https://github.com/shinagawa-web/gomarklint/releases/latest).

```sh
# macOS / Linux
tar -xzf gomarklint_Darwin_x86_64.tar.gz
mv gomarklint /usr/local/bin/

# Windows (PowerShell)
Expand-Archive gomarklint_Windows_x86_64.zip
mv gomarklint.exe C:\Windows\System32\
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
