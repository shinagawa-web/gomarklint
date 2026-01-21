package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

// TestE2E_ValidMarkdown tests linting a valid markdown file
func TestE2E_ValidMarkdown(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/valid.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("expected exit 0, got error: %v\noutput: %s", err, output)
	}

	// Verify output indicates no issues found
	if !bytes.Contains(output, []byte("No issues found")) {
		t.Errorf("expected 'No issues found' in output, got: %s", output)
	}
}

// TestE2E_InvalidHeadingLevel tests linting a file with heading level errors
func TestE2E_InvalidHeadingLevel(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/invalid_heading_level.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	output, _ := cmd.CombinedOutput()

	// Verify error message is detected in output
	if !bytes.Contains(output, []byte("First heading should be level 2")) {
		t.Errorf("expected 'First heading should be level 2' error in output, got: %s", output)
	}

	// Verify issues count is shown
	if !bytes.Contains(output, []byte("1 issues found")) {
		t.Errorf("expected '1 issues found' in output, got: %s", output)
	}
}

// TestE2E_DuplicateHeadings tests linting a file with duplicate headings
func TestE2E_DuplicateHeadings(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/duplicate_headings.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	output, _ := cmd.CombinedOutput()

	// Verify duplicate heading error is detected in output
	if !bytes.Contains(output, []byte("duplicate heading")) {
		t.Errorf("expected 'duplicate heading' error in output, got: %s", output)
	}

	// Verify issues count is shown
	if !bytes.Contains(output, []byte("1 issues found")) {
		t.Errorf("expected '1 issues found' in output, got: %s", output)
	}
}

// TestE2E_MultipleBlankLines tests linting a file with multiple consecutive blank lines
func TestE2E_MultipleBlankLines(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/multiple_blank_lines.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	output, _ := cmd.CombinedOutput()

	// Verify multiple blank lines error is detected in output
	if !bytes.Contains(output, []byte("Multiple consecutive blank lines")) {
		t.Errorf("expected 'Multiple consecutive blank lines' error in output, got: %s", output)
	}

	// Verify issues count is shown
	if !bytes.Contains(output, []byte("1 issues found")) {
		t.Errorf("expected '1 issues found' in output, got: %s", output)
	}
}

// TestE2E_CLIFlagsOverrideConfig tests that CLI flags override config file settings
func TestE2E_CLIFlagsOverrideConfig(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/heading_level_one.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	// Override minHeadingLevel from 2 to 1 via CLI flag
	cmd.Args = append(cmd.Args, "--min-heading", "1")
	output, _ := cmd.CombinedOutput()

	// With --min-heading 1, H1 should be allowed and no error should occur
	if !bytes.Contains(output, []byte("No issues found")) {
		t.Errorf("expected 'No issues found' when --min-heading 1 overrides config, got: %s", output)
	}

	// Should not contain heading level error
	if bytes.Contains(output, []byte("First heading should be level")) {
		t.Errorf("expected no heading level error with --min-heading 1, got: %s", output)
	}
}

// TestE2E_DisableRuleViaFlag tests that rules can be disabled via CLI flags
func TestE2E_DisableRuleViaFlag(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/duplicate_headings.md"

	// Check if binary exists
	if _, err := os.Stat("./" + binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryPath, err)
	}

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	cmd.Args = append(cmd.Args, "--config", ".gomarklint.json")
	// Disable duplicate heading check via CLI flag
	cmd.Args = append(cmd.Args, "--enable-duplicate-heading-check=false")
	output, _ := cmd.CombinedOutput()

	// With duplicate heading check disabled, no error should occur
	if !bytes.Contains(output, []byte("No issues found")) {
		t.Errorf("expected 'No issues found' when duplicate heading check is disabled, got: %s", output)
	}

	// Should not contain duplicate heading error
	if bytes.Contains(output, []byte("duplicate heading")) {
		t.Errorf("expected no duplicate heading error when disabled, got: %s", output)
	}
}
