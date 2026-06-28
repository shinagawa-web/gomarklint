package preprocess

import "testing"

func TestIndentColumns(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantCols     int
		wantFirstIdx int
	}{
		{"no indent", "text", 0, 0},
		{"three spaces", "   x", 3, 3},
		{"four spaces", "    x", 4, 4},
		// A tab advances to the next multiple of four columns (CommonMark tab
		// stops), so a single leading tab is four columns of indentation.
		{"single tab", "\tx", 4, 1},
		{"two spaces then tab", "  \tx", 4, 3},
		{"tab then text", "\t\tx", 8, 2},
		{"blank line", "", 0, 0},
		{"whitespace-only line", "   ", 3, 3},
		{"tab-only line", "\t", 4, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, firstIdx := indentColumns(tt.line)
			if cols != tt.wantCols || firstIdx != tt.wantFirstIdx {
				t.Errorf("indentColumns(%q) = (%d, %d), want (%d, %d)",
					tt.line, cols, firstIdx, tt.wantCols, tt.wantFirstIdx)
			}
		})
	}
}
