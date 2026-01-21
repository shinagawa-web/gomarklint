package e2e

import (
	"os/exec"
	"testing"
)

// TestE2E_ValidMarkdown tests linting a valid markdown file
func TestE2E_ValidMarkdown(t *testing.T) {
	binaryPath := "gomarklint-e2e-test"
	fixturePath := "fixtures/valid.md"

	cmd := exec.Command("./" + binaryPath)
	cmd.Args = append(cmd.Args, fixturePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("expected exit 0, got error: %v\noutput: %s", err, output)
	}
}
