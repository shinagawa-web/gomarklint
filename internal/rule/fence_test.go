package rule

import "testing"

func TestIsClosingFence(t *testing.T) {
	tests := []struct {
		name       string
		trimmed    string
		openMarker string
		want       bool
	}{
		{"exact backtick match", "```", "```", true},
		{"exact tilde match", "~~~", "~~~", true},
		{"longer closing backtick", "````", "```", true},
		{"longer closing tilde", "~~~~", "~~~", true},
		{"much longer closing", "``````", "```", true},
		{"shorter closing rejected", "```", "````", false},
		{"trailing whitespace allowed", "```  ", "```", true},
		{"trailing text rejected", "``` foo", "```", false},
		{"mismatched fence char", "~~~", "```", false},
		{"mismatched fence char reverse", "```", "~~~", false},
		{"empty trimmed", "", "```", false},
		{"empty marker", "```", "", false},
		{"both empty", "", "", false},
		{"4-backtick exact", "````", "````", true},
		{"4-backtick longer close", "`````", "````", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsClosingFence(tt.trimmed, tt.openMarker)
			if got != tt.want {
				t.Errorf("IsClosingFence(%q, %q) = %v, want %v", tt.trimmed, tt.openMarker, got, tt.want)
			}
		})
	}
}
