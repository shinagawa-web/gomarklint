---
title: "GitHub Actions"
weight: 7
---

# GitHub Actions Integration

You can use gomarklint in your CI workflows using the official [GitHub Action](https://github.com/marketplace/actions/gomarklint-markdown-linter).

> **Note:** When using `gomarklint` in GitHub Actions, you must first create a `.gomarklint.json` configuration file in your repository root.
> Generate a default config with: `gomarklint init`

## Example: `.github/workflows/lint.yml`

```yml
name: Lint Markdown

on:
  push:
    paths:
      - '**/*.md'
  pull_request:
    paths:
      - '**/*.md'

jobs:
  markdown-lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run gomarklint Action
        uses: shinagawa-web/gomarklint-action@v1
```
