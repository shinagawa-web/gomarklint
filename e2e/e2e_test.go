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
