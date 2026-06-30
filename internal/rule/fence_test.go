package rule

import "testing"

// TestOpeningFenceMarker exercises openingFenceMarker directly. The style-family
// migration to preprocess.Context (#337 Phase 3) removed its incidental coverage
// from those rules' tests; its only remaining caller, CheckFencedCodeLanguage,
// feeds it pre-validated fence openers from ctx.FenceSpans() and so never hits
// the guard branches.
func TestOpeningFenceMarker(t *testing.T) {
	tests := []struct {
		name    string
		trimmed string
		want    string
	}{
		{"backtick fence", "```", "```"},
		{"tilde fence", "~~~", "~~~"},
		{"longer backtick run", "````", "````"},
		{"fence with language", "```go", "```"},
		{"tilde with language", "~~~bash", "~~~"},
		{"too short", "``", ""},
		{"empty", "", ""},
		{"not a fence char", "---", ""},
		{"second char mismatch", "`x`", ""},
		{"third char mismatch", "``x", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := openingFenceMarker(tt.trimmed)
			if got != tt.want {
				t.Errorf("openingFenceMarker(%q) = %q, want %q", tt.trimmed, got, tt.want)
			}
		})
	}
}
