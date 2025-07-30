# gomarklint

![Test](https://github.com/shinagawa-web/gomarklint/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/gh/shinagawa-web/gomarklint/graph/badge.svg?token=5MGCYZZY7S)](https://codecov.io/gh/shinagawa-web/gomarklint)
[![Go Report Card](https://goreportcard.com/badge/github.com/shinagawa-web/gomarklint)](https://goreportcard.com/report/github.com/shinagawa-web/gomarklint)
[![Go Reference](https://pkg.go.dev/badge/github.com/shinagawa-web/gomarklint.svg)](https://pkg.go.dev/github.com/shinagawa-web/gomarklint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> A fast and lightweight Markdown linter written in Go.

**gomarklint** checks your Markdown files for common issues such as heading structure problems, trailing blank lines, unclosed code blocks, and more. Designed to be minimal, fast, and CI-friendly.

---

## ✨ Features

- ✅ Lint individual `.md` files or entire directories
- ✅ Checks for heading level consistency (`# → ## → ###`)
- ✅ Detects duplicate headings (case-insensitive, trims trailing spaces)
- ✅ Detects missing trailing blank lines
- ✅ Detects unclosed code blocks
- ✅ Ignores YAML frontmatter correctly when linting
- ✅ Detects broken external links (e.g. `[text](https://...)`, `https://...`) with `--enable-link-check`
- ✅ Supports config file (`.gomarklint.json`) to store default options
- ✅ Supports ignore patterns (e.g. `**/CHANGELOG.md`) via config file
- ✅ Supports structured JSON output via `--output=json`
- ⚡️ Blazing fast — 157 files and 52,000+ lines scanned in under 50ms
- 🐢 External link checking is slower (e.g. ~160s for 157 files), but optional and off by default


📝 **Note:** By default, `gomarklint` assumes heading levels start from `##` (H2), not `#` (H1), to align with common blog and static site conventions.


## 📋 Example Output

### Text Output

```bash
❯ go run main.go testdata/sample_links.md

Errors in testdata/sample_links.md:
  testdata/sample_links.md:1: First heading should be level 2 (found level 1)
  testdata/sample_links.md:4: Link unreachable: https://httpstat.us/404
  testdata/sample_links.md:12: Link unreachable: http://localhost-test:3001
  testdata/sample_links.md:16: duplicate heading: "overview"
  testdata/sample_links.md:18: image with empty alt text


✖ 5 issues found
✓ Checked 1 file(s), 19 line(s) in 757ms
```

### JSON Output

```bash
❯ gomarklint testdata/sample_links.md --output=json
```

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


## 📦 Installation (for local linting)

### Option 1: Pre-built Binaries (Recommended for most users)

You can download the latest binary from GitHub Releases.

```bash
# Example (Linux, x86_64)
curl -L -o gomarklint.tar.gz https://github.com/shinagawa-web/gomarklint/releases/latest/download/gomarklint_Linux_x86_64.tar.gz
tar -xzf gomarklint.tar.gz
mv gomarklint /usr/local/bin/
```

- Binaries for macOS, Linux, and Windows are available.
- Windows users: download the `.zip` version and extract it manually
- All binaries are statically compiled with `CGO_ENABLED=0`, so no external dependencies are required.


### Option 2: go install (for Go users)

```bash
go install github.com/shinagawa-web/gomarklint@latest
```

### Option 3: Clone and run locally

```bash
git clone https://github.com/shinagawa-web/gomarklint.git
cd gomarklint
go run main.go ./README.md
```

## 🚀 Usage

```bash
gomarklint ./posts --min-heading 2
gomarklint ./posts ./docs
gomarklint ./content --ignore CHANGELOG.md --json
gomarklint ./docs --enable-link-check
```

Options:

- `--min-heading` — Set the minimum heading level to expect. Defaults to `2` (i.e. `##`), which aligns with common blogging/static site practices.
- `--enable-link-check` — Check for broken external links (http/https) such as [text](https://...), ![alt](https://...), or bare URLs. Only runs when explicitly enabled.
  - Example: `[text](https://...)`, `![img](https://...)`, or bare URLs
> 🕒 Note: With `--enable-link-check` enabled, performance depends on network conditions.
> For example, checking 157 files (~52,000 lines) with link validation may take ~100s.

- `--skip-link-patterns` — (optional) One or more regular expressions to exclude specific URLs from link checking. Useful for skipping `localhost`, internal domains, etc.
  - Example: `--skip-link-patterns localhost --skip-link-patterns ^https://internal\.example\.com`
- `--output` — Set output format. Accepts `"text"` (default) or `"json"`.  Use `"json"` to generate structured output for CI tools, scripts, etc.
  - Example: `--output=json`

## ⚙️ Configuration File

`gomarklint` supports configuration via a `.gomarklint.json` file.

By default, if the file exists in the current directory, it will be loaded automatically. You can also specify a custom path using the `--config` flag.

```json
{
  "minHeadingLevel": 2,
  "checkLinks": true,
  "skipLinkPatterns": [
    "localhost",
    "example.com"
  ],
  "ignore": [],
  "output": "text"
}
```

- CLI flags override values in the config file.
- Unknown fields in the JSON will cause an error (strict validation).
- Valid values: `"text"` (default) or `"json"`

you can generate a default config file using:

```bash
gomarklint init
```

## ⚡️ Performance Tips

`gomarklint` is built for speed.  
For example, scanning **157 files and 52,000+ lines** takes under **50ms** when external link checking is disabled.

However, when using `--enable-link-check`, performance may slow down because:

- External links require real **HTTP requests**
- Network latency, timeouts, or retries can significantly impact speed
- More links = more waiting

### ✅ Recommended usage

Use `--enable-link-check` only when necessary, such as:

- Nightly CI runs
- Pre-release validation
- Verifying newly added content

If you want lightning-fast feedback while editing, omit `--enable-link-check`.

> ⏱️ **Fastest mode:**  
> `gomarklint ./content` → ✅ completes in milliseconds!

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
        with:
          args: markdown
```

### Args

- args: Specify the directory or files to lint. For example: `.` or `docs` or `README.md.`

> By default, the action runs gomarklint using the latest release binary.

## 🛣 Roadmap

v0.1.0
- [x] Basic CLI setup with Cobra
- [x] Lint single file
- [x] Check heading level jumps
- [x] Check missing final blank lines
- [x] Check unclosed code blocks

v0.2.0
- [x] Support multiple files and directories
- [x] Output file name and line number
- [x] Recursively search .md files
- [x] Frontmatter support

v0.2.1
- [x] Extract `http`/`https` URLs from:
  - Inline links: `[text](https://example.com)`
  - Image links: `![alt](https://example.com/image.png)`
  - Bare URLs: `https://example.com/path`
- [x] Perform HTTP `HEAD` or fallback `GET` request to validate links
- [x] Report links that return 4xx or 5xx status codes
- [x] Set request timeout (default: 5 seconds)
- [x] Include automated tests (with mock server for consistent results)
- [x] Show file name and line number for each broken link
- [x] `--enable-link-check` flag to enable external link checking
- [x] Skip external link checking unless `--enable-link-check` is specified

v0.2.2
- [x] Support `--skip-link-patterns` to exclude certain domains from link checking
- [x] Ignore links inside fenced code blocks (```...```)
- [x] Remove No issues found 🎉 message (consider --quiet flag)

v0.3.0 - Configuration File Support

- [x] Define `Config` struct to represent configuration options
- [x] Add `--config` flag to specify config file path (default: `.gomarklint.json`)
- [x] Load configuration via `os.ReadFile` and `json.Unmarshal`
- [x] Determine priority between flags and config file (e.g., flags override config or vice versa)
- [x] Handle missing config file gracefully and apply default values
- [x] Add `gomarklint init` subcommand to generate a default `.gomarklint.json` file

v0.4.0
- [x] Add rules: duplicate headings, empty alt text
- [x] Add --ignore flag
- [x] Add --json output option

v0.5.0
- [x] Cross-platform binaries via goreleaser
- [x] GitHub Actions support

v1.0.0
- [ ] At least 5 rules with test coverage
- [ ] Stable CLI interface
- [ ] Prebuilt binaries for macOS/Linux/Windows
- [ ] Clear README and blog post with real usage examples

## 📁 Project Structure

```
gomarklint/
├── cmd/             # CLI entrypoint (Cobra)
├── internal/
│   ├── rule/        # Individual lint rules
│   └── parser/      # Markdown parsing logic
├── testdata/        # Sample Markdown files
├── main.go
└── README.md
```

## 📁 Path Handling

When specifying files or directories, `gomarklint` will:

- Recursively search `.md` files using `filepath.WalkDir`
- Ignore hidden directories like `.git/`
- Skip symbolic links
- Report all files, regardless of `.gitignore`
- Silently skip missing files (`os.IsNotExist`)

## 🤝 Contributing

Issues, suggestions, and PRs are welcome!
This project is just getting started and will evolve step by step.

## 📜 License

MIT License
