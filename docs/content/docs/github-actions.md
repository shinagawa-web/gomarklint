---
title: "GitHub Actions"
weight: 8
---

# GitHub Actions Integration

You can use gomarklint in your CI workflows using the official [GitHub Action](https://github.com/marketplace/actions/gomarklint-markdown-linter).

> **Note:** It is recommended that you create a `.gomarklint.json` configuration file in your repository root before using `gomarklint` in GitHub Actions. If no configuration file is present, gomarklint will run with its default settings.
> You can generate a starter config with: `gomarklint init`

## Quick Start

```yml
# .github/workflows/docs-lint.yml
name: Lint Markdown

on:
  pull_request:
    paths:
      - '**/*.md'

jobs:
  docs-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - uses: shinagawa-web/gomarklint-action@v1
```

## PR Comment

Post lint results as a comment on pull requests. The comment is automatically updated on subsequent runs, avoiding duplicates.

```yml
- uses: shinagawa-web/gomarklint-action@v1
  with:
    comment-on-pr: 'true'
    github-token: ${{ secrets.GITHUB_TOKEN }}
```

> **Note:** The job needs `pull-requests: write` permission.

```yml
permissions:
  contents: read
  pull-requests: write
```

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `args` | No | `''` | Arguments to pass to gomarklint |
| `comment-on-pr` | No | `'false'` | Post lint results as a PR comment |
| `github-token` | No | `${{ github.token }}` | GitHub token for posting PR comments |
