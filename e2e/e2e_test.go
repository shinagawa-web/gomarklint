package e2e

import (
	"bytes"
	"encoding/json"
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
	assertOutputContains(t, output, "First heading should be level 2")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_DuplicateHeadings tests linting a file with duplicate headings
func TestE2E_DuplicateHeadings(t *testing.T) {
	output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "duplicate heading")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_MultipleBlankLines tests linting a file with multiple consecutive blank lines
func TestE2E_MultipleBlankLines(t *testing.T) {
	output := runTest(t, "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
	assertOutputContains(t, output, "Multiple consecutive blank lines")
	assertOutputContains(t, output, "1 issues found")
}

// TestE2E_CLIFlagsOverrideConfig tests that CLI flags override config file settings
func TestE2E_CLIFlagsOverrideConfig(t *testing.T) {
	output := runTest(t, "fixtures/heading_level_one.md", "--config", ".gomarklint.json", "--min-heading", "1")
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "First heading should be level")
}

// TestE2E_DisableRuleViaFlag tests that rules can be disabled via CLI flags
func TestE2E_DisableRuleViaFlag(t *testing.T) {
	output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json", "--enable-duplicate-heading-check=false")
	assertOutputContains(t, output, "No issues found")
	assertOutputNotContains(t, output, "duplicate heading")
}

// TestE2E_TextFormat tests that text output format shows readable error messages
func TestE2E_TextFormat(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "text")
	assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
	assertOutputContains(t, output, "First heading should be level 2")
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
}

// TestE2E_JSONFormat tests that JSON output format produces valid JSON with correct file paths and line numbers
func TestE2E_JSONFormat(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "json")

	// Parse JSON to verify it's valid
	var result map[string]any
	err := json.Unmarshal(output, &result)
	if err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, output)
	}

	// Verify top-level fields exist
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
}

// TestE2E_MultipleFiles tests that multiple files can be specified as arguments and all are checked
func TestE2E_MultipleFiles(t *testing.T) {
	output := runTest(t, "fixtures/valid.md", "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")

	// invalid_heading_level.md should have heading level error
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md")
	assertOutputContains(t, output, "First heading should be level 2")

	// Should report multiple files checked
	assertOutputContains(t, output, "Checked 2 file(s)")
}

// TestE2E_DirectoryRecursion tests that directories are processed recursively
func TestE2E_DirectoryRecursion(t *testing.T) {
	output := runTest(t, "fixtures", "--config", ".gomarklint.json")

	// Should process multiple files from fixtures directory
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md")
	assertOutputContains(t, output, "fixtures/duplicate_headings.md")
	assertOutputContains(t, output, "fixtures/multiple_blank_lines.md")

	// Should verify specific error messages for each file
	assertOutputContains(t, output, "First heading should be level 2")
	assertOutputContains(t, output, "duplicate heading")
	assertOutputContains(t, output, "Multiple consecutive blank lines")

	// Should report total issues from all files
	assertOutputContains(t, output, "5 issues found")

	// Should report files checked (heading_level_one.md also has error, empty.md missing final blank line)
	assertOutputContains(t, output, "Checked 7 file(s)")
}

// TestE2E_ErrorsFromAllFiles tests that errors from multiple files are reported correctly
func TestE2E_ErrorsFromAllFiles(t *testing.T) {
	output := runTest(t, "fixtures/invalid_heading_level.md", "fixtures/duplicate_headings.md", "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")

	// invalid_heading_level.md has heading level error
	assertOutputContains(t, output, "fixtures/invalid_heading_level.md")
	assertOutputContains(t, output, "First heading should be level 2")

	// duplicate_headings.md has duplicate heading error
	assertOutputContains(t, output, "fixtures/duplicate_headings.md")
	assertOutputContains(t, output, "duplicate heading")

	// multiple_blank_lines.md has blank line error
	assertOutputContains(t, output, "fixtures/multiple_blank_lines.md")
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
