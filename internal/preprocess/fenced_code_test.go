package preprocess

import "testing"

func TestOpeningFenceMarker(t *testing.T) {
	tests := []struct {
		name    string
		trimmed string
		want    string
	}{
		{"triple backtick", "```", "```"},
		{"backtick with info string", "```go", "```"},
		{"longer backtick run", "````", "````"},
		{"triple tilde", "~~~", "~~~"},
		{"tilde with info string", "~~~ruby", "~~~"},
		{"two backticks is not a fence", "``", ""},
		{"single backtick", "`", ""},
		{"not a fence", "plain text", ""},
		{"empty", "", ""},
		{"mixed run is not a fence", "`~`", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := openingFenceMarker(tt.trimmed); got != tt.want {
				t.Errorf("openingFenceMarker(%q) = %q, want %q", tt.trimmed, got, tt.want)
			}
		})
	}
}

func TestIsClosingFence(t *testing.T) {
	tests := []struct {
		name       string
		trimmed    string
		openMarker string
		want       bool
	}{
		// CommonMark: a closing fence must use the same character and be at
		// least as long as the opener.
		{"exact match", "```", "```", true},
		{"longer closer", "````", "```", true},
		{"shorter closer does not close", "```", "````", false},
		{"different char does not close", "~~~", "```", false},
		{"trailing whitespace allowed", "```   ", "```", true},
		{"trailing text does not close", "``` go", "```", false},
		{"empty opener", "```", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isClosingFence(tt.trimmed, tt.openMarker); got != tt.want {
				t.Errorf("isClosingFence(%q, %q) = %v, want %v", tt.trimmed, tt.openMarker, got, tt.want)
			}
		})
	}
}
