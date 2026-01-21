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
