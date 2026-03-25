package rule

import "testing"

func TestIsClosingFence(t *testing.T) {
	tests := []struct {
		name        string
		trimmed     string
		fenceMarker string
		want        bool
	}{
		{
			name:        "exact match backtick",
			trimmed:     "```",
			fenceMarker: "```",
			want:        true,
		},
		{
			name:        "longer closing fence (CommonMark)",
			trimmed:     "````",
			fenceMarker: "```",
			want:        true,
		},
		{
			name:        "exact match tilde",
			trimmed:     "~~~",
			fenceMarker: "~~~",
			want:        true,
		},
		{
			name:        "longer tilde closing fence",
			trimmed:     "~~~~",
			fenceMarker: "~~~",
			want:        true,
		},
		{
			name:        "trailing whitespace allowed",
			trimmed:     "```  ",
			fenceMarker: "```",
			want:        true,
		},
		{
			name:        "shorter than opening is not a closing fence",
			trimmed:     "``",
			fenceMarker: "```",
			want:        false,
		},
		{
			name:        "wrong fence character",
			trimmed:     "~~~",
			fenceMarker: "```",
			want:        false,
		},
		{
			name:        "fence with language is not a closing fence",
			trimmed:     "```go",
			fenceMarker: "```",
			want:        false,
		},
		{
			name:        "empty line is not a closing fence",
			trimmed:     "",
			fenceMarker: "```",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClosingFence(tt.trimmed, tt.fenceMarker)
			if got != tt.want {
				t.Errorf("isClosingFence(%q, %q) = %v, want %v", tt.trimmed, tt.fenceMarker, got, tt.want)
			}
		})
	}
}
