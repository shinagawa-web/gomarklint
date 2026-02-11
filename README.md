# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> A fast, opinionated Markdown linter for engineering teams. Built in Go, designed for CI.

- Catch broken links and headings before your docs ship.
- Enforce predictable structure (no more “why is this H4 under H2?”).
- Output that’s friendly for both humans and machines (JSON).

## Why

Docs break quietly and trust erodes loudly.
gomarklint focuses on reproducible rules that prevent “small but costly” failures:

- Heading hierarchies that drift during edits
- Duplicate headings that break anchor links
- Subtle dead links (including internal anchors)
- Large repos where “one-off checks” don’t scale

> Goal: treat documentation quality like code quality—fast feedback locally, strict in CI, zero drama.

## ✨ Features

- **⚡️ Blazingly fast**: Process **100,000+ lines in ~170ms** (structural checks only, M4 Mac)
- Recursive .md search (multi-file & multi-directory)
- Frontmatter-aware parsing (YAML/TOML ignored when needed)
- File name & line number in diagnostics
- Human-readable and JSON outputs
- Fast single-binary CLI (Go), ideal for CI/CD
- Rules with clear rationales (see below)

Planned/ongoing:

- Severity levels per rule
- Customizable rule enable/disable
- VS Code extension for in-editor feedback

## Quick Start

```sh
# install (choose one)
go install github.com/shinagawa-web/gomarklint@latest

# or clone and build manually
git clone https://github.com/shinagawa-web/gomarklint
cd gomarklint
make build   # or: go build ./cmd/gomarklint
```

### 1) Initialize config (optional but recommended)

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

### 2) Run it

```sh
# lint current directory recursively
gomarklint ./...

# lint specific targets
gomarklint docs README.md internal/handbook
```

Exit code is non-zero if any violations are found, zero otherwise.

### 3) JSON output (for CI / tooling)

```sh
gomarklint ./... --output json
```

## Rules (current)

`gomarklint` currently runs the following checks (ordered as executed):

| Rule key                       | What it detects                                        | Notes / Options                                                                                        |
| ------------------------------ | ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| `final-blank-line`             | Missing final blank line at EOF                        | Always on                                                                                              |
| `unclosed-code-block`          | Unclosed fenced code blocks (````` / `~~~`)            | Always on                                                                                              |
| `empty-alt-text`               | Image syntax with an empty alt text                    | Always on                                                                                              |
| `heading-level`                | Invalid heading level progression (e.g., H2 → H4 skip) | Toggle: `--enable-heading-level-check` (default **on**) / `--min-heading` (default **2**)              |
| `duplicate-heading`            | Duplicate headings within one file                     | Toggle: `--enable-duplicate-heading-check` (default **on**)                                            |
| `no-multiple-blank-lines`      | Multiple consecutive blank lines                       | Toggle: `--enable-no-multiple-blank-lines-check` (default **on**)                                      |
| `external-link`                | External links that fail validation                    | Toggle: `--enable-link-check` (default **off**). Skips URLs that match `--skip-link-patterns` (regex). |
| `no-setext-headings`           | Setext heading used instead of ATX style               | Toggle: `--enable-no-setext-headings-check` (default **on**)                                           |

Execution details:

- Files/dirs are expanded with ignore patterns from config (see Configuration).
- Per-file issues are sorted by line asc before printing.
- Line count is computed as \n count + 1 for reporting.

## CLI

```sh
gomarklint [files or directories] [flags]
```

If no paths are given, the tool will:

- Use `include` from `.gomarklint.json` if present, otherwise error out with
“please provide a markdown file or directory (or set 'include' in .gomarklint.json)”.

### Flags

| Flag                               | Type             | Default            | Description                                                                            |
| ---------------------------------- | ---------------- | ------------------ | -------------------------------------------------------------------------------------- |
| `--config`                         | string           | `.gomarklint.json` | Path to config file. Loaded if the file exists.                                        |
| `--min-heading`                    | int              | `2`                | Minimum heading level considered by the heading-level rule.                            |
| `--enable-link-check`              | bool             | `false`            | Enable external link checking.                                                         |
| `--enable-heading-level-check`     | bool             | `true`             | Enable heading level validation.                                                       |
| `--enable-duplicate-heading-check` | bool             | `true`             | Enable duplicate heading detection.                                                    |
| `--skip-link-patterns`             | string[] (regex) | `[]`               | Regex patterns; matching URLs are skipped by link check. Can be passed multiple times. |
| `--output`                         | `text` | `json`  | `text`             | Output format. Any other value is rejected.                                            |

Notes:

- Flags override config values when explicitly provided.
- Paths are expanded (globs/dirs) and filtered by ignore (from config).
- Exit behavior: the command returns a non-nil error (non-zero exit), zero otherwise.

## Configuration

A JSON config is read from the path given by --config (defaults to .gomarklint.json) if the file exists. Example:

```json
{
  "include": ["docs", "README.md"],
  "ignore": ["node_modules", "vendor"],
  "outputFormat": "text",
  "minHeadingLevel": 2,
  "enableLinkCheck": false,
  "enableHeadingLevelCheck": true,
  "enableDuplicateHeadingCheck": true,
  "skipLinkPatterns": [
    "^https://localhost(:[0-9]+)?/",
    "example\\.com"
  ]
}
```

Field effects:

- If CLI flags are set, they take precedence over config.
- If no CLI paths are provided, include (when present) becomes the target set.

## Output

### Human-readable (`--output text`, default)

- Prints grouped file sections only when a file has issues:

```sh
❯ gomarklint testdata/sample_links.md

Errors in testdata/sample_links.md:
  testdata/sample_links.md:1: First heading should be level 2 (found level 1)
  testdata/sample_links.md:4: Link unreachable: https://httpstat.us/404
  testdata/sample_links.md:12: Link unreachable: http://localhost-test:3001
  testdata/sample_links.md:16: duplicate heading: "overview"
  testdata/sample_links.md:18: image with empty alt text


✖ 5 issues found
✓ Checked 1 file(s), 19 line(s) in 757ms
```

- Summary and timing:
  - If issues: `✖ N issues found`
  - If none: `✔ No issues found`
  - Always prints: `Checked <files>, <lines> in <Xms|Ys>` with colored ticks.

### JSON (`--output json`)

```json
{
  "files": 1,
  "lines": 19,
  "errors": 5,
  "elapsed_ms": 790,
  "details": {
    "testdata/sample_links.md": [
      {
        "File": "testdata/sample_links.md",
        "Line": 1,
        "Message": "First heading should be level 2 (found level 1)"
      },
      {
        "File": "testdata/sample_links.md",
        "Line": 4,
        "Message": "Link unreachable: https://httpstat.us/404"
      },
      {
        "File": "testdata/sample_links.md",
        "Line": 12,
        "Message": "Link unreachable: http://localhost-test:3001"
      },
      {
        "File": "testdata/sample_links.md",
        "Line": 16,
        "Message": "duplicate heading: \"overview\""
      },
      {
        "File": "testdata/sample_links.md",
        "Line": 18,
        "Message": "image with empty alt text"
      }
    ]
  }
}
```

- details maps file path → list of issues (`file`, `line`, `message`).
- elapsed_ms is total wall time for the run.

## ⚡️ Performance

`gomarklint` is built for speed, with optimizations for both file parsing and external link validation.

**Structural checks** (headings, code blocks, etc.):
- Scanning **185 files and 104,000+ lines** takes under **60ms**

**External link checking** (`--enable-link-check`):
- Optimized concurrent validation with intelligent batching
- **~2,000 external links** validated in **under 10 seconds**
- Significantly faster than traditional sequential HTTP checks

### ✅ Recommended usage

**For rapid local feedback:**
- Run without `--enable-link-check` → completes in milliseconds
- Perfect for catching structural issues while editing

**For comprehensive validation:**
- Enable `--enable-link-check` for:
  - Nightly CI runs
  - Pre-release validation
  - Verifying newly added content
- Performance remains practical even at scale

> ⏱️ **TL;DR:**  
> Fast enough for local dev (no link check), robust enough for CI (with link check).

## 📊 Benchmarking

`gomarklint` includes comprehensive benchmarks to track performance and prevent regressions.

### Running Benchmarks Locally

```bash
# Run all benchmarks
make bench

# Run benchmarks for a specific package
go test -bench=. ./internal/rule/

# Run with memory profiling
go test -bench=. -benchmem ./...
```

### Benchmark Coverage

Benchmarks are available for the main linting workflows:
- **Lint rules**: Each rule has dedicated benchmarks (e.g., `BenchmarkCheckHeadingLevel`)
- **Full linting**: End-to-end benchmarks for running gomarklint across files (see `cmd/root_bench_test.go`)

### CI Integration

Pull requests automatically run benchmark comparisons against the main branch:
- Shows performance differences for each benchmarked function
- Highlights regressions with visual indicators (✅/⚠️/❌)
- Results are posted as PR comments for easy review

The benchmark workflow ensures performance remains stable across code changes.

## 🧪 GitHub Actions Integration

You can use gomarklint in your CI workflows using the official [GitHub Action](https://github.com/marketplace/actions/gomarklint-markdown-linter):

> ⚠️ Note:
> When using `gomarklint` in GitHub Actions, you must first create a `.gomarklint.json` configuration file in your repository root.
> This ensures all options are explicitly defined and reproducible in CI environments.

You can generate a default config with:

```bash
gomarklint init
```

Example: .github/workflows/lint.yml

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

## 🛣️ Roadmap (Post v1.0.0)

### ✅ Core Quality & Rule Expansion

- [ ] `max-line-length`: Enforce maximum line width
- [x] `no-multiple-consecutive-blank-lines`: Disallow multiple blank lines
- [ ] `image-alt-text` improvements: Enforce alt text style and length
- [ ] Rule severity levels (e.g. `warning`, `error`)

### 🧩 Extensibility

- [ ] Plugin system for custom rules (via Go interface or external binary)
- [ ] Allow disabling specific rules via inline comments (e.g. `<!-- gomarklint-disable -->`)

### 🧪 Testing & Stability

- [ ] Snapshot testing support for easier rule verification
- [ ] Regression test suite for real-world Markdown samples

### 🛠️ Developer UX

- [ ] VS Code extension using gomarklint core
- [ ] Interactive mode (e.g. prompt to fix or explain errors)
- [ ] File caching for faster repeated linting

### 📦 Ecosystem & CI

- [x] GitHub Actions integration
- [x] Prebuilt binaries via `goreleaser` (macOS/Linux/Windows)
- [ ] Homebrew formula
- [ ] Docker image (e.g. `ghcr.io/shinagawa-web/gomarklint`)

### 🌍 Internationalization

- [ ] Localized messages (e.g. Japanese, Spanish)
- [ ] Rule messages with IDs and documentation links

---

**Feel free to suggest more ideas by opening an issue or discussion on GitHub!**

## 📁 Project Structure

```
gomarklint/
├── cmd/                    # CLI commands
│   ├── init.go            # Configuration initialization
│   ├── root.go            # Root command and CLI orchestration
│   └── root_bench_test.go # Benchmark tests for CLI
├── internal/
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config struct and defaults
│   │   ├── config_test.go
│   │   ├── load.go        # Configuration file loading
│   │   ├── merge.go       # Config merging and flag handling
│   │   └── merge_test.go
│   ├── file/              # File system operations
│   │   ├── expand.go      # File expansion and glob pattern matching
│   │   ├── expand_test.go
│   │   ├── pathutil.go    # Path utilities
│   │   ├── pathutil_test.go
│   │   ├── reader.go      # File reading with frontmatter handling
│   │   └── reader_test.go
│   ├── linter/            # Core linting logic
│   │   ├── linter.go      # Linter implementation with concurrent processing
│   │   └── linter_test.go
│   ├── output/            # Output formatting
│   │   ├── formatter.go   # Formatter interface
│   │   ├── json.go        # JSON output formatter
│   │   ├── json_test.go
│   │   ├── text.go        # Text output formatter
│   │   ├── text_test.go
│   │   └── testutil_test.go
│   ├── rule/              # Lint rules implementation
│   │   ├── code_block.go
│   │   ├── code_block_test.go
│   │   ├── code_block_bench_test.go
│   │   ├── duplicate_headings.go
│   │   ├── duplicate_headings_test.go
│   │   ├── duplicate_headings_bench_test.go
│   │   ├── empty_alt_text.go
│   │   ├── empty_alt_text_test.go
│   │   ├── empty_alt_text_bench_test.go
│   │   ├── external_link.go
│   │   ├── external_link_test.go
│   │   ├── external_link_bench_test.go
│   │   ├── final_blank_line.go
│   │   ├── final_blank_line_test.go
│   │   ├── final_blank_line_bench_test.go
│   │   ├── heading_level.go
│   │   ├── heading_level_test.go
│   │   ├── heading_level_bench_test.go
│   │   ├── no_multiple_blank_lines.go
│   │   ├── no_multiple_blank_lines_test.go
│   │   ├── no_multiple_blank_lines_bench_test.go
│   │   ├── setext_headings.go
│   │   └── setext_headings_test.go
│   └── testutil/          # Testing utilities
│       ├── path.go
│       └── path_test.go
├── e2e/                   # End-to-end tests
│   ├── e2e_test.go
│   ├── fixtures/          # Test fixture markdown files
│   └── .gomarklint.json
├── testdata/              # Unit test fixtures
├── main.go               # Application entry point
└── doc.go                # Package documentation
```

## 📁 Path Handling

When specifying files or directories, `gomarklint` will:

- Recursively search `.md` files using `filepath.WalkDir`
- Ignore hidden directories like `.git/`
- Skip symbolic links
- Report all files, regardless of `.gitignore`
- Silently skip missing files (`os.IsNotExist`)

## 🛠 Local Development

To set up a local development environment for `gomarklint`:

### Using Make (Recommended)

```bash
# Show all available commands
make help

# Build the binary
make build

# Run unit tests
make test

# Run end-to-end tests
make test-e2e

# Run all tests (unit + E2E)
make test-all

# Run tests with coverage report
make test-coverage

# Lint the included sample files in ./testdata
make lint

# Lint the repo's README
make lint-self

# Run gomarklint with custom arguments
make run-dev ARGS="README.md"

# Generate a default .gomarklint.json
make init

# Clean build artifacts
make clean
```

### Testing Strategy

#### Unit Tests
Unit tests for individual rules and utilities are located in `*_test.go` files alongside the code they test:
- `internal/rule/*_test.go` — Test individual lint rules
- `internal/linter/*_test.go` — Test core linting logic
- `internal/file/*_test.go` — Test file operations and path utilities
- `internal/config/*_test.go` — Test configuration loading and merging
- `internal/output/*_test.go` — Test output formatters

Run with: `make test`

#### End-to-End Tests
E2E tests verify the complete CLI behavior by running the compiled binary against fixture files:
- Located in `e2e/e2e_test.go`
- Test fixtures in `e2e/fixtures/` (Markdown files with various rule violations)
- Tests are organized into logical categories:
  - Basic Functionality: Individual rule detection (heading levels, duplicates, blank lines, code blocks, alt text, external links)
  - Configuration: CLI flag overrides and rule disabling
  - Output Formats: Text and JSON output validation
  - Multiple Files: Multi-file and directory recursion handling
  - Edge Cases: Non-existent files, invalid configs, empty files, frontmatter handling, multiple violations in single file

Run with: `make test-e2e`

Notes:
- `go run .` uses the local source directly, so you don't need to `go install` during development.
- When adding new CLI flags or config fields, confirm they appear in `--help` and the generated `.gomarklint.json`.
- Tests should remain fast and self-contained — contributions that break this will be rejected.
- E2E tests may take longer due to external link validation; use `enableLinkCheck: false` in test config when testing rules unrelated to links.

## 🤝 Contributing

Issues, suggestions, and PRs are welcome!
Before submitting a pull request, please follow the guidelines below to ensure a smooth review process.

Requirements:

- Go version: Go `1.22+` (latest stable recommended)

## 📜 License

MIT License
