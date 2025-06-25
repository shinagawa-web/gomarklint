package rule

import (
	"os"
	"testing"

	"github.com/shinagawa-web/gomarklint/internal/testutil"
)

func TestCheckUnclosedCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{"valid", "testdata/code_block/valid.md", false},
		{"unclosed", "testdata/code_block/unclosed.md", true},
		{"frontmatter_unclosed", "testdata/code_block/frontmatter_unclosed.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := testutil.GetTestFilePath(tt.filepath)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}
			errs := CheckUnclosedCodeBlocks(path, string(data))
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, errs)
			}
		})
	}
}
