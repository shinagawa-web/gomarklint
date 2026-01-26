# GitHub Copilot Instructions for gomarklint

## Project Overview

`gomarklint` is a Markdown linter written in Go that checks for various issues in Markdown files, including:
- Heading level consistency
- External link validation
- Duplicate headings
- Code block formatting
- Empty alt text in images
- Final blank line requirements

## Project Structure

```
gomarklint/
├── cmd/                    # CLI commands
│   ├── init.go            # Configuration initialization
│   └── root.go            # Root command and main logic
├── e2e/                    # End-to-end tests
│   ├── e2e_test.go        # E2E test cases
│   ├── fixtures/          # Test fixture markdown files
│   └── .gomarklint.json   # Config for E2E tests
├── internal/
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config struct and defaults
│   │   ├── config_test.go
│   │   └── load.go        # Configuration loading
│   ├── parser/            # Markdown parsing utilities
│   │   ├── expand.go      # File expansion logic
│   │   ├── expand_test.go
│   │   ├── external_link.go # External link handling
│   │   ├── external_link_test.go
│   │   ├── markdown.go    # Core markdown parsing
│   │   ├── markdown_test.go
│   │   ├── strip_frontmatter.go # Frontmatter removal
│   │   └── strip_frontmatter_test.go
│   ├── rule/              # Lint rules implementation
│   │   ├── code_block.go
│   │   ├── code_block_test.go
│   │   ├── duplicate_headings.go
│   │   ├── duplicate_headings_test.go
│   │   ├── empty_alt_text.go
│   │   ├── empty_alt_text_test.go
│   │   ├── external_link.go
│   │   ├── external_link_test.go
│   │   ├── external_link_internal_test.go
│   │   ├── final_blank_line.go
│   │   ├── final_blank_line_test.go
│   │   ├── heading_level.go
│   │   ├── heading_level_test.go
│   │   ├── no_multiple_blank_lines.go
│   │   └── no_multiple_blank_lines_test.go
│   ├── testutil/          # Testing utilities
│   │   ├── path.go
│   │   └── path_test.go
│   └── util/              # Common utilities
│       ├── pathutil.go
│       └── pathutil_test.go
├── testdata/              # Unit test fixtures
├── main.go               # Application entry point
├── doc.go                # Package documentation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile              # Build and test targets
└── README.md             # Project documentation
```

## Development Guidelines

### Configuration

- Config struct is in `internal/config/config.go`
- All config fields should have JSON tags for serialization
- Default values are defined in the `Default()` function
- New configuration options should be added to both the struct and defaults

### Rules Implementation

- Each lint rule is in its own file under `internal/rule/`
- Rules should follow the pattern: `Check{RuleName}(path, content string, ...) []LintError`
- Include comprehensive tests for each rule
- Rules should be configurable via the Config struct when applicable

### Testing

- Follow Go testing conventions with `_test.go` files
- Use table-driven tests where appropriate
- Test both positive and negative cases

### CLI Commands

- Main CLI logic is in `cmd/root.go`
- Command flags should correspond to config options
- Use cobra framework for CLI implementation
- Support both config file and command line flag configuration

### Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Add comments for exported functions and complex logic
- **All comments must be in English** - no Japanese or other non-English comments

## Common Tasks

### Adding a New Lint Rule

1. Create new file in `internal/rule/`
2. Implement the check function returning `[]LintError`
3. Add unit tests in corresponding `_test.go` file
4. Add config option if needed in `internal/config/config.go`
5. Integrate into main checking logic in `cmd/root.go`
6. Add E2E test case in `e2e/e2e_test.go` with test fixture in `e2e/fixtures/` if applicable

### Adding Configuration Options

1. Add field to `Config` struct with JSON tag
2. Update `Default()` function with default value
3. Add command line flag in `cmd/root.go` if needed
4. Update configuration loading logic if required

### Option 3: Clone and run locally

```bash
git clone https://github.com/shinagawa-web/gomarklint.git
cd gomarklint
go run main.go ./README.md
```

### Running Commands

- Build: `go build -o gomarklint .`
- Run directly: `go run . [command] [flags]`
- Initialize config: `go run . init`
- Run linter: `go run . [files...]`

## Key Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/bmatcuk/doublestar` - Glob pattern matching
- Standard library packages for file I/O, regex, HTTP

## 🛠 Local Development

To set up a local development environment for `gomarklint`:

```bash
# Run unit tests only
make test

# Run end-to-end tests
make test-e2e

# Run all tests (unit + E2E)
make test-all

# Build the binary
make build

# Show CLI help from local source
go run . --help

# Generate a default .gomarklint.json (from your local build)
go run . init
```

### Testing Strategy

- **Unit Tests**: Tests for individual rules and utilities are in `*_test.go` files alongside the code
- **E2E Tests**: Integration tests in `e2e/e2e_test.go` test the full CLI behavior against fixture files in `e2e/fixtures/`
- Run `make build-e2e` to build the binary for E2E tests (automatically done by `make test-e2e`)

Notes:
- `go run .` uses the local source directly, so you don't need to `go install` during development.
- When adding new CLI flags or config fields, confirm they appear in `--help` and the generated `.gomarklint.json`.
- Tests should remain fast and self-contained — contributions that break this will be rejected.
- When adding new rules or CLI flags, add corresponding E2E tests in `e2e/e2e_test.go` and test fixtures in `e2e/fixtures/`

## Notes for AI Assistance

- When modifying config, always update both the struct and Default() function
- New rules should be added to the main checking logic in collectErrors()
- Follow existing patterns for error handling and return types
- Prefer using the existing test utilities in internal/testutil/
- Consider backwards compatibility when making config changes
