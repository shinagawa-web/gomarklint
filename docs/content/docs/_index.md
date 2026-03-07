---
title: "Quick Start"
weight: 1
---

# Quick Start

## Install

```sh
# install via go install
go install github.com/shinagawa-web/gomarklint@latest

# or clone and build manually
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
