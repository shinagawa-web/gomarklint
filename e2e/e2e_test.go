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

// TestE2E organizes all E2E tests into logical categories
func TestE2E(t *testing.T) {
	t.Run("Basic Functionality", func(t *testing.T) {
		t.Run("ValidMarkdown", func(t *testing.T) {
			output := runTest(t, "fixtures/valid.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "No issues found")
		})

		t.Run("InvalidHeadingLevel", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")
			assertOutputContains(t, output, "found level 1")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("DuplicateHeadings", func(t *testing.T) {
			output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
			assertOutputContains(t, output, "duplicate heading")
			assertOutputContains(t, output, "\"section one\"")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("MultipleBlankLines", func(t *testing.T) {
			output := runTest(t, "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
			assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
			assertOutputContains(t, output, "Multiple consecutive blank lines")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("UnclosedCodeBlock", func(t *testing.T) {
			output := runTest(t, "fixtures/unclosed_code_block.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
			assertOutputContains(t, output, "fixtures/unclosed_code_block.md:5:")
			assertOutputContains(t, output, "Unclosed code block")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("EmptyAltText", func(t *testing.T) {
			output := runTest(t, "fixtures/empty_alt_text.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
			assertOutputContains(t, output, "fixtures/empty_alt_text.md:5:")
			assertOutputContains(t, output, "image with empty alt text")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("InvalidExternalLink", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
			assertOutputContains(t, output, "Link unreachable")
			assertOutputContains(t, output, "https://this-domain-definitely-does-not-exist-12345.com")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})
	})

	t.Run("Configuration", func(t *testing.T) {
		t.Run("CLIFlagsOverrideConfig", func(t *testing.T) {
			output := runTest(t, "fixtures/heading_level_one.md", "--config", ".gomarklint.json", "--min-heading", "1")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "First heading should be level")
			assertOutputNotContains(t, output, "Errors")
		})

		t.Run("DisableRuleViaFlag", func(t *testing.T) {
			output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json", "--enable-duplicate-heading-check=false")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "duplicate heading")
			assertOutputNotContains(t, output, "Errors")
		})
	})

	t.Run("Output Formats", func(t *testing.T) {
		t.Run("TextFormat", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "text")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("JSONFormat", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "json")
			assertOutputContains(t, output, `"files"`)
			assertOutputContains(t, output, `"errors"`)
			assertOutputContains(t, output, `"details"`)
			assertOutputContains(t, output, `"elapsed_ms"`)
			assertOutputContains(t, output, `"File": "fixtures/invalid_heading_level.md"`)
			assertOutputContains(t, output, `"Line": 1`)
			assertOutputContains(t, output, `"Message": "First heading should be level 2`)
			assertOutputContains(t, output, `{`)
			assertOutputContains(t, output, `}`)
		})
	})

	t.Run("Multiple Files", func(t *testing.T) {
		t.Run("MultipleFiles", func(t *testing.T) {
			output := runTest(t, "fixtures/valid.md", "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")
			assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")
			assertOutputContains(t, output, "Checked 2 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("DirectoryRecursion", func(t *testing.T) {
			output := runTest(t, "fixtures", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2 (found level 1)")
			assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
			assertOutputContains(t, output, "duplicate heading: \"section one\"")
			assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
			assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
			assertOutputContains(t, output, "Multiple consecutive blank lines")
			assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
			assertOutputContains(t, output, "fixtures/unclosed_code_block.md:5:")
			assertOutputContains(t, output, "Unclosed code block")
			assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
			assertOutputContains(t, output, "fixtures/empty_alt_text.md:5:")
			assertOutputContains(t, output, "image with empty alt text")
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
			assertOutputContains(t, output, "Link unreachable: https://this-domain-definitely-does-not-exist-12345.com")
			assertOutputContains(t, output, "Errors in fixtures/heading_level_one.md:")
			assertOutputContains(t, output, "fixtures/heading_level_one.md:1:")
			assertOutputContains(t, output, "Errors in fixtures/empty.md:")
			assertOutputContains(t, output, "fixtures/empty.md:1:")
			assertOutputContains(t, output, "Missing final blank line")
			assertOutputContains(t, output, "Errors in fixtures/multiple_violations.md:")
			assertOutputContains(t, output, "fixtures/multiple_violations.md:1:")
			assertOutputContains(t, output, "13 issues found")
			assertOutputContains(t, output, "Checked 11 file(s)")
			assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
			assertOutputNotContains(t, output, "Errors in fixtures/with_frontmatter.md")
		})

		t.Run("ErrorsFromAllFiles", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_heading_level.md", "fixtures/duplicate_headings.md", "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")
			assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:9:")
			assertOutputContains(t, output, "duplicate heading")
			assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
			assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
			assertOutputContains(t, output, "Multiple consecutive blank lines")
			assertOutputContains(t, output, "Checked 3 file(s)")
			assertOutputContains(t, output, "3 issues found")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("NonExistentFile", func(t *testing.T) {
			output := runTest(t, "fixtures/nonexistent.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Checked 0 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "Errors")
		})

		t.Run("InvalidConfigFile", func(t *testing.T) {
			output, err := runTestWithCmd(t, "fixtures/valid.md", "--config", "invalid.json")
			if err == nil {
				t.Errorf("expected error for invalid config file, but command succeeded")
			}
			errorOutput := string(output)
			if !bytes.Contains(output, []byte("error")) && !bytes.Contains(output, []byte("invalid")) {
				t.Errorf("expected error message about invalid config, got: %s", errorOutput)
			}
		})

		t.Run("EmptyFile", func(t *testing.T) {
			output := runTest(t, "fixtures/empty.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/empty.md:")
			assertOutputContains(t, output, "fixtures/empty.md:1:")
			assertOutputContains(t, output, "Missing final blank line")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("FilesWithFrontmatter", func(t *testing.T) {
			output := runTest(t, "fixtures/with_frontmatter.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "Errors")
		})

		t.Run("MultipleViolationsInSingleFile", func(t *testing.T) {
			output, _ := runTestWithCmd(t, "fixtures/multiple_violations.md", "--config", ".gomarklint.json")

			// Should return output showing errors (exit code behavior may vary)
			// Just verify errors are reported

			// All errors should be reported
			assertOutputContains(t, output, "Errors in fixtures/multiple_violations.md:")

			// Heading level error (first heading is level 1, should be level 2)
			assertOutputContains(t, output, "fixtures/multiple_violations.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")

			// Multiple blank lines error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:5:")
			assertOutputContains(t, output, "Multiple consecutive blank lines")

			// Duplicate heading error (line 12)
			assertOutputContains(t, output, "fixtures/multiple_violations.md:12:")
			assertOutputContains(t, output, "duplicate heading")
			assertOutputContains(t, output, "section one")

			// Empty alt text error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:16:")
			assertOutputContains(t, output, "image with empty alt text")

			// Unclosed code block error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:20:")
			assertOutputContains(t, output, "Unclosed code block")

			// File should have been checked
			assertOutputContains(t, output, "Checked 1 file(s)")

			// All 5 errors should be reported
			assertOutputContains(t, output, "5 issues found")
		})
	})
}
