package util

import "testing"

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{
			name:     "exact match",
			path:     "drafts/test.md",
			patterns: []string{"drafts/test.md"},
			want:     true,
		},
		{
			name:     "wildcard match",
			path:     "drafts/test.md",
			patterns: []string{"drafts/*.md"},
			want:     true,
		},
		{
			name:     "recursive wildcard match",
			path:     "docs/api/test.md",
			patterns: []string{"**/*.md"},
			want:     true,
		},
		{
			name:     "no match",
			path:     "content/post.md",
			patterns: []string{"drafts/*.md"},
			want:     false,
		},
		{
			name:     "empty patterns",
			path:     "test.md",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "directory match",
			path:     "docs/intro.md",
			patterns: []string{"docs/**"},
			want:     true,
		},
		{
			name:     "partial match should fail",
			path:     "docs/readme.md",
			patterns: []string{"doc/*.md"}, // typo
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIgnore(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q, %v) = %v; want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}
