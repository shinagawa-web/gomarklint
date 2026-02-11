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
│   ├── root.go            # Root command and CLI orchestration
│   └── root_bench_test.go # Benchmark tests for CLI
├── e2e/                    # End-to-end tests
│   ├── e2e_test.go        # E2E test cases
│   ├── fixtures/          # Test fixture markdown files
│   ├── invalid.json       # Invalid config for testing
│   └── .gomarklint.json   # Config for E2E tests
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
│   │   ├── external_link_internal_test.go
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
- Configuration loading is in `internal/config/load.go`
- Configuration merging with CLI flags is in `internal/config/merge.go`
- New configuration options should be added to the struct, defaults, and merge logic

### Linting Logic

- Core linting logic is in `internal/linter/linter.go`
- The `Linter` struct encapsulates configuration and state (e.g., URL cache)
- `Run()` method processes files concurrently using goroutines
- `LintContent()` method lints string content without file I/O (useful for benchmarks)
- Frontmatter is stripped automatically before applying rules
- All rules are applied in `collectErrors()` method

### Rules Implementation

- Each lint rule is in its own file under `internal/rule/`
- Rules should follow the pattern: `Check{RuleName}(path, content string, ...) []LintError`
- Include comprehensive tests for each rule
- Rules should be configurable via the Config struct when applicable
- Benchmark tests should be added for performance-critical rules

### Output Formatting

- Output formatters are in `internal/output/`
- Implement the `Formatter` interface for new output formats
- `TextFormatter` provides human-readable output with color support
- `JSONFormatter` provides machine-readable structured output
- Formatters receive a `Result` from the linter and format errors accordingly

### Testing

- Follow Go testing conventions with `_test.go` files
- Use table-driven tests where appropriate
- Test both positive and negative cases

### CLI Commands

- Main CLI orchestration is in `cmd/root.go`
- Heavy lifting delegated to `internal/linter/` and `internal/output/`
- Command flags should correspond to config options
- Use cobra framework for CLI implementation
- Flag merging with config file is handled by `internal/config/merge.go`
- Error handling distinguishes between lint violations (`ErrLintViolations`) and real errors

### File Operations

- File system operations are in `internal/file/`
- `expand.go` handles glob pattern matching and file discovery
- `reader.go` handles file reading with automatic frontmatter stripping
- `pathutil.go` provides path normalization utilities

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
5. Integrate into main checking logic in `internal/linter/linter.go` (`collectErrors` method)
6. Add E2E test case in `e2e/e2e_test.go` with test fixture in `e2e/fixtures/` if applicable

### Adding Configuration Options

1. Add field to `Config` struct with JSON tag in `internal/config/config.go`
2. Update `Default()` function with default value
3. Add flag merging logic in `internal/config/merge.go` if needed
4. Add command line flag in `cmd/root.go` if needed
5. Update configuration validation in `internal/config/merge.go` if required

### Adding Output Formats

1. Create new formatter in `internal/output/`
2. Implement the `Formatter` interface
3. Add tests for the new formatter
4. Update `cmd/root.go` to support the new format option
5. Add to output format validation in `internal/config/merge.go`

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
- New rules should be added to the main checking logic in `internal/linter/linter.go` (`collectErrors()` method)
- Follow existing patterns for error handling and return types
- Prefer using the existing test utilities in `internal/testutil/`
- Consider backwards compatibility when making config changes
- The main.go handles exit codes; distinguish between `ErrLintViolations` and real errors
