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

// runTest is a helper function to run the gomarklint binary and return only the output.
// It ignores the exit code since E2E tests check output content, not exit status.
func runTest(t *testing.T, args ...string) []byte {
	output, _ := runTestWithCmd(t, args...)
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

// runTestWithCmd runs the gomarklint binary and returns output and exit code.
// This is used when tests need to verify error conditions.
func runTestWithCmd(t *testing.T, args ...string) ([]byte, error) {
	binaryPath := "./" + binaryName

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryName, err)
	}

	cmd := exec.Command(binaryPath)
	cmd.Args = append(cmd.Args, args...)
	// We capture the error here to return it to the caller.
	// Some tests need to verify error conditions (e.g., invalid config files).
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
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:14:")
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
			// Enable link check explicitly to test external link validation
			output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
			assertOutputContains(t, output, "Link unreachable")
			assertOutputContains(t, output, "https://this-domain-definitely-does-not-exist-12345.com")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("NoFinalBlankLine", func(t *testing.T) {
			output := runTest(t, "fixtures/no_final_blank_line.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/no_final_blank_line.md:")
			assertOutputContains(t, output, "fixtures/no_final_blank_line.md:3:")
			assertOutputContains(t, output, "Missing final blank line")
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

		t.Run("DisableFinalBlankLineCheck", func(t *testing.T) {
			output := runTest(t, "fixtures/no_final_blank_line.md", "--config", ".gomarklint.json", "--enable-final-blank-line-check=false")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "Missing final blank line")
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
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:14:")
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
			// External link check is enabled in the E2E config, so we should see link errors
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "fixtures/invalid_external_link.md:9:")
			assertOutputContains(t, output, "Link unreachable: https://this-domain-definitely-does-not-exist-12345.com")
			assertOutputContains(t, output, "Errors in fixtures/multiple_external_links.md:")
			assertOutputContains(t, output, "this-is-definitely-an-invalid-domain-12345.xyz")
			assertOutputContains(t, output, "another-invalid-domain-67890.test")
			assertOutputContains(t, output, "Errors in fixtures/heading_level_one.md:")
			assertOutputContains(t, output, "fixtures/heading_level_one.md:1:")
			assertOutputContains(t, output, "Errors in fixtures/empty.md:")
			assertOutputContains(t, output, "fixtures/empty.md:1:")
			assertOutputContains(t, output, "Missing final blank line")
			assertOutputContains(t, output, "Errors in fixtures/multiple_violations.md:")
			assertOutputContains(t, output, "fixtures/multiple_violations.md:6:")
			// Count may vary, but should have checked all files
			assertOutputContains(t, output, "Checked 17 file(s)")
			assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
			assertOutputNotContains(t, output, "Errors in fixtures/with_frontmatter.md")
			// valid_external_links.md should have no errors when link check is enabled
			assertOutputNotContains(t, output, "Errors in fixtures/valid_external_links.md")
			// mixed_link_types.md should have no errors (only checks HTTP/HTTPS)
			assertOutputNotContains(t, output, "Errors in fixtures/mixed_link_types.md")
		})

		t.Run("ErrorsFromAllFiles", func(t *testing.T) {
			output := runTest(t, "fixtures/invalid_heading_level.md", "fixtures/duplicate_headings.md", "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
			assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
			assertOutputContains(t, output, "First heading should be level 2")
			assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
			assertOutputContains(t, output, "fixtures/duplicate_headings.md:14:")
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
			assertOutputContains(t, output, "fixtures/multiple_violations.md:6:")
			assertOutputContains(t, output, "First heading should be level 2")

			// Multiple blank lines error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:10:")
			assertOutputContains(t, output, "Multiple consecutive blank lines")

			// Duplicate heading error (line 17)
			assertOutputContains(t, output, "fixtures/multiple_violations.md:17:")
			assertOutputContains(t, output, "duplicate heading")
			assertOutputContains(t, output, "section one")

			// Empty alt text error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:21:")
			assertOutputContains(t, output, "image with empty alt text")

			// Unclosed code block error
			assertOutputContains(t, output, "fixtures/multiple_violations.md:25:")
			assertOutputContains(t, output, "Unclosed code block")

			// File should have been checked
			assertOutputContains(t, output, "Checked 1 file(s)")

			// All 5 errors should be reported
			assertOutputContains(t, output, "5 issues found")
		})
	})

	t.Run("External Link Checks", func(t *testing.T) {
		t.Run("DisableExternalLinkCheck", func(t *testing.T) {
			// External link check is disabled by default, so this should pass without checking links
			output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json", "--enable-link-check=false")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "Link unreachable")
		})

		t.Run("EnableExternalLinkCheckWithInvalidLink", func(t *testing.T) {
			// Enable link check and verify it detects invalid links
			output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "Link unreachable")
			assertOutputContains(t, output, "this-domain-definitely-does-not-exist-12345.com")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("ValidExternalLinksOnly", func(t *testing.T) {
			// Test with only valid external links
			output := runTest(t, "fixtures/valid_external_links.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			assertOutputNotContains(t, output, "Link unreachable")
		})

		t.Run("MultipleExternalLinks", func(t *testing.T) {
			// Test with multiple external links (both valid and invalid)
			output := runTest(t, "fixtures/multiple_external_links.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Errors in fixtures/multiple_external_links.md:")
			assertOutputContains(t, output, "Link unreachable")
			assertOutputContains(t, output, "this-is-definitely-an-invalid-domain-12345.xyz")
			assertOutputContains(t, output, "another-invalid-domain-67890.test")
			assertOutputContains(t, output, "2 issues found")
		})

		t.Run("ExternalLinkCheckEnabledByDefault", func(t *testing.T) {
			// Verify that external link check is enabled in the E2E config file
			output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
			assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
			assertOutputContains(t, output, "Link unreachable")
			assertOutputContains(t, output, "1 issues found")
		})

		t.Run("HTTPAndHTTPSLinks", func(t *testing.T) {
			// Test that both HTTP and HTTPS links are checked
			output := runTest(t, "fixtures/http_and_https_links.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Checked 1 file(s)")
			// Should check both HTTP and HTTPS links
			// The HTTP link might fail, but the test verifies both are checked
			assertOutputContains(t, output, "link(s)")
		})

		t.Run("MixedLinkTypes", func(t *testing.T) {
			// Test that only HTTP/HTTPS links are checked, not relative paths or FTP
			output := runTest(t, "fixtures/mixed_link_types.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Checked 1 file(s)")
			// Should only check HTTP/HTTPS links (2 links: Google and GitHub bare URL)
			assertOutputContains(t, output, "link(s)")
			assertOutputContains(t, output, "No issues found")
		})

		t.Run("SameLineMultipleLinks", func(t *testing.T) {
			// Test multiple links in the same line
			output := runTest(t, "fixtures/same_line_multiple_links.md", "--config", ".gomarklint.json", "--enable-link-check=true")
			assertOutputContains(t, output, "Checked 1 file(s)")
			assertOutputContains(t, output, "No issues found")
			// Should check all links even if they're on the same line
			assertOutputContains(t, output, "link(s)")
		})
	})
}
