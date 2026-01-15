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
├── internal/
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config struct and defaults
│   │   ├── config_test.go
│   │   └── load.go        # Configuration loading
│   ├── parser/            # Markdown parsing utilities
│   │   ├── expand.go      # File expansion logic
│   │   ├── external_link.go # External link handling
│   │   ├── markdown.go    # Core markdown parsing
│   │   └── strip_frontmatter.go # Frontmatter removal
│   ├── rule/              # Lint rules implementation
│   │   ├── code_block.go
│   │   ├── duplicate_headings.go
│   │   ├── empty_alt_text.go
│   │   ├── external_link.go
│   │   ├── final_blank_line.go
│   │   └── heading_level.go
│   ├── testutil/          # Testing utilities
│   └── util/              # Common utilities
├── testdata/              # Test fixtures
├── main.go               # Application entry point
└── doc.go                # Package documentation
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

- Use `testdata/` directory for test fixtures
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
3. Add tests in corresponding `_test.go` file
4. Add config option if needed in `internal/config/config.go`
5. Integrate into main checking logic in `cmd/root.go`

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
- `gopkg.in/yaml.v3` - YAML parsing (for frontmatter)
- Standard library packages for file I/O, regex, HTTP

## 🛠 Local Development

To set up a local development environment for `gomarklint`:

```bash
# Run all tests
go test ./...

# Show CLI help from local source
go run . --help

# Generate a default .gomarklint.json (from your local build)
go run . init

# Lint the included sample files in ./testdata
go run . testdata
```

Notes:
- `go run .` uses the local source directly, so you don't need to `go install` during development.
- When adding new CLI flags or config fields, confirm they appear in `--help` and the generated `.gomarklint.json`.
- Tests should remain fast and self-contained — contributions that break this will be rejected.

## Notes for AI Assistance

- When modifying config, always update both the struct and Default() function
- New rules should be added to the main checking logic in collectErrors()
- Follow existing patterns for error handling and return types
- Prefer using the existing test utilities in internal/testutil/
- Consider backwards compatibility when making config changes
