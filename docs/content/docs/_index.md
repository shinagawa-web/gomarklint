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
make build   # or: go build ./cmd/gomarklint
```

## 1) Initialize config (optional but recommended)

```sh
gomarklint init
```

This creates `.gomarklint.json` with sensible defaults:

```json
{
  "include": ["."],
  "ignore": ["node_modules", "vendor"],
  "minHeadingLevel": 2,
  "enableHeadingLevelCheck": true,
  "enableDuplicateHeadingCheck": true,
  "enableLinkCheck": false,
  "enableNoSetextHeadingsCheck": true,
  "skipLinkPatterns": [],
  "outputFormat": "text"
}
```

You can edit it anytime — CLI flags override config values.

## 2) Run it

```sh
# lint current directory recursively
gomarklint ./...

# lint specific targets
gomarklint docs README.md internal/handbook
```

Exit code is non-zero if any violations are found, zero otherwise.

## 3) JSON output (for CI / tooling)

```sh
gomarklint ./... --output json
```
