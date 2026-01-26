package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

const (
	binaryName = "gomarklint-e2e-test"
)

// runTest is a helper function to run the gomarklint binary with given arguments and return the output
func runTest(t *testing.T, args ...string) []byte {
	binaryPath := "./" + binaryName

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryName, err)
	}

	cmd := exec.Command(binaryPath)
	cmd.Args = append(cmd.Args, args...)
	output, _ := cmd.CombinedOutput()
	return output
}

// assertOutputContains checks if output contains the expected string
func assertOutputContains(t *testing.T, output []byte, expected string) {
	if !bytes.Contains(output, []byte(expected)) {
		t.Errorf("expected output to contain %q, got: %s", expected, output)
	}
}

// assertOutputNotContains checks if output does not contain the unexpected string
func assertOutputNotContains(t *testing.T, output []byte, unexpected string) {
	if bytes.Contains(output, []byte(unexpected)) {
		t.Errorf("expected output to NOT contain %q, got: %s", unexpected, output)
	}
}

// runTestWithCmd runs the gomarklint binary and returns output and a flag indicating if binary executed
func runTestWithCmd(t *testing.T, args ...string) ([]byte, error) {
	binaryPath := "./" + binaryName

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryName, err)
	}

	cmd := exec.Command(binaryPath)
	cmd.Args = append(cmd.Args, args...)
	output, err := cmd.CombinedOutput()
	return output, err
}

// TestE2E_ValidMarkdown tests linting a valid markdown file
func TestE2E_ValidMarkdown(t *testing.T) {
	output := runTest(t, "fixtures/valid.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "No issues found")
}

// TestE2E_InvalidHeadingLevel tests linting a file with heading level errors
func TestE2E_InvalidHeadingLevel(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
	assertOutputContains(t, output, "First heading should be level 2")
	assertOutputContains(t, output, "found level 1")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_DuplicateHeadings tests linting a file with duplicate headings
func TestE2E_DuplicateHeadings(t *testing.T) {
	output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
	assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
	assertOutputContains(t, output, "duplicate heading")
	assertOutputContains(t, output, "\"section one\"")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_MultipleBlankLines tests linting a file with multiple consecutive blank lines
func TestE2E_MultipleBlankLines(t *testing.T) {
	output := runTest(t, "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
	assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
	assertOutputContains(t, output, "Multiple consecutive blank lines")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_CLIFlagsOverrideConfig tests that CLI flags override config file settings
func TestE2E_CLIFlagsOverrideConfig(t *testing.T) {
	output := runTest(t, "fixtures/heading_level_one.md", "--config", ".gomarklint.json", "--min-heading", "1")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "First heading should be level")
	assertOutputNotContains(t, output, "Errors")
}

// TestE2E_DisableRuleViaFlag tests that rules can be disabled via CLI flags
func TestE2E_DisableRuleViaFlag(t *testing.T) {
	output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json", "--enable-duplicate-heading-check=false")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "duplicate heading")
	assertOutputNotContains(t, output, "Errors")
}

// TestE2E_TextFormat tests that text output format shows readable error messages
func TestE2E_TextFormat(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "text")
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
	assertOutputContains(t, output, "First heading should be level 2")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_JSONFormat tests that JSON output format produces valid JSON with correct file paths and line numbers
func TestE2E_JSONFormat(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "json")

	// Verify top-level fields exist in JSON
	assertOutputContains(t, output, `"files"`)
	assertOutputContains(t, output, `"errors"`)
	assertOutputContains(t, output, `"details"`)
	assertOutputContains(t, output, `"elapsed_ms"`)

	// Verify correct file path is in JSON
	assertOutputContains(t, output, `"File": "fixtures/invalid_heading_level.md"`)

	// Verify line number is in JSON
	assertOutputContains(t, output, `"Line": 1`)

	// Verify error message is in JSON
	assertOutputContains(t, output, `"Message": "First heading should be level 2`)

	// Verify JSON structure with opening brace
	assertOutputContains(t, output, `{`)
	assertOutputContains(t, output, `}`)
}

// TestE2E_MultipleFiles tests that multiple files can be specified as arguments and all are checked
func TestE2E_MultipleFiles(t *testing.T) {
	output := runTest(t, "fixtures/valid.md", "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")

	// valid.md should have no errors
	assertOutputNotContains(t, output, "Errors in fixtures/valid.md")

	// invalid_heading_level.md should have heading level error
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
	assertOutputContains(t, output, "First heading should be level 2")

	// Should report multiple files checked and one issue
	assertOutputContains(t, output, "Checked 2 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_DirectoryRecursion tests that directories are processed recursively
func TestE2E_DirectoryRecursion(t *testing.T) {
	output := runTest(t, "fixtures", "--config", ".gomarklint.json")

	// Verify each file with errors is properly reported with detailed information

	// invalid_heading_level.md errors
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
	assertOutputContains(t, output, "First heading should be level 2 (found level 1)")

	// duplicate_headings.md errors
	assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
	assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
	assertOutputContains(t, output, "duplicate heading: \"section one\"")

	// multiple_blank_lines.md errors
	assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
	assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
	assertOutputContains(t, output, "Multiple consecutive blank lines")

	// unclosed_code_block.md errors
	assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
	assertOutputContains(t, output, "fixtures/unclosed_code_block.md:5:")
	assertOutputContains(t, output, "Unclosed code block")

	// empty_alt_text.md errors
	assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
	assertOutputContains(t, output, "fixtures/empty_alt_text.md:5:")
	assertOutputContains(t, output, "image with empty alt text")

	// invalid_external_link.md errors
	assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
	assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
	assertOutputContains(t, output, "Link unreachable: https://this-domain-definitely-does-not-exist-12345.com")

	// heading_level_one.md errors (additional file with heading level issue)
	assertOutputContains(t, output, "Errors in fixtures/heading_level_one.md:")
	assertOutputContains(t, output, "fixtures/heading_level_one.md:1:")

	// empty.md errors (missing final blank line)
	assertOutputContains(t, output, "Errors in fixtures/empty.md:")
	assertOutputContains(t, output, "fixtures/empty.md:1:")
	assertOutputContains(t, output, "Missing final blank line")

	// Summary statistics
	assertOutputContains(t, output, "8 issues found")
	assertOutputContains(t, output, "Checked 10 file(s)")

	// Verify valid.md and with_frontmatter.md have no errors
	assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
	assertOutputNotContains(t, output, "Errors in fixtures/with_frontmatter.md")
}

// TestE2E_ErrorsFromAllFiles tests that errors from multiple files are reported correctly
func TestE2E_ErrorsFromAllFiles(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "fixtures/duplicate_headings.md", "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")

	// invalid_heading_level.md has heading level error
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
	assertOutputContains(t, output, "First heading should be level 2")

	// duplicate_headings.md has duplicate heading error
	assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
	assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
	assertOutputContains(t, output, "duplicate heading")

	// multiple_blank_lines.md has blank line error
	assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
	assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
	assertOutputContains(t, output, "Multiple consecutive blank lines")

	// Should report multiple errors
	assertOutputContains(t, output, "Checked 3 file(s)")
	assertOutputContains(t, output, "3 issues found")
}

// TestE2E_NonExistentFile tests that non-existent files are handled gracefully
func TestE2E_NonExistentFile(t *testing.T) {
	output := runTest(t, "fixtures/nonexistent.md", "--config", ".gomarklint.json")

	// Non-existent file is skipped, resulting in no files checked
	assertOutputContains(t, output, "Checked 0 file(s)")
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "Errors")
}

// TestE2E_InvalidConfigFile tests that invalid config files produce appropriate errors
func TestE2E_InvalidConfigFile(t *testing.T) {
	output, err := runTestWithCmd(t, "fixtures/valid.md", "--config", "invalid.json")

	// Should have error
	if err == nil {
		t.Errorf("expected error for invalid config file, but command succeeded")
	}

	// Should contain error message about config
	errorOutput := string(output)
	if !bytes.Contains(output, []byte("error")) && !bytes.Contains(output, []byte("invalid")) {
		t.Errorf("expected error message about invalid config, got: %s", errorOutput)
	}
}

// TestE2E_EmptyFile tests that empty files are processed
func TestE2E_EmptyFile(t *testing.T) {
	output := runTest(t, "fixtures/empty.md", "--config", ".gomarklint.json")

	// Empty file triggers final blank line check
	assertOutputContains(t, output, "Errors in fixtures/empty.md:")
	assertOutputContains(t, output, "fixtures/empty.md:1:")
	assertOutputContains(t, output, "Missing final blank line")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_FilesWithFrontmatter tests that files with frontmatter are correctly processed
func TestE2E_FilesWithFrontmatter(t *testing.T) {
	output := runTest(t, "fixtures/with_frontmatter.md", "--config", ".gomarklint.json")

	// File with frontmatter should be stripped and processed (has H2 headings which are valid)
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "Errors")
}

// TestE2E_UnclosedCodeBlock tests that files with unclosed code blocks are detected
func TestE2E_UnclosedCodeBlock(t *testing.T) {
	output := runTest(t, "fixtures/unclosed_code_block.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
	assertOutputContains(t, output, "fixtures/unclosed_code_block.md:5:")
	assertOutputContains(t, output, "Unclosed code block")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_EmptyAltText tests that images with empty alt text are detected
func TestE2E_EmptyAltText(t *testing.T) {
	output := runTest(t, "fixtures/empty_alt_text.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
	assertOutputContains(t, output, "fixtures/empty_alt_text.md:5:")
	assertOutputContains(t, output, "image with empty alt text")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_InvalidExternalLink tests that invalid/unreachable external links are detected
func TestE2E_InvalidExternalLink(t *testing.T) {
	output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
	assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
	assertOutputContains(t, output, "Link unreachable")
	assertOutputContains(t, output, "https://this-domain-definitely-does-not-exist-12345.com")
	assertOutputContains(t, output, "Checked 1 file(s)")
	assertOutputContains(t, output, "1 issues found")
}
