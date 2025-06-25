package parser

import (
	"testing"
)

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBody string
		wantSkip int
	}{
		{
			name: "with frontmatter",
			input: `---
title: "Test"
date: 2025-01-01
---

# Hello`,
			wantBody: `# Hello`,
			wantSkip: 5,
		},
		{
			name:     "no frontmatter",
			input:    `# Hello`,
			wantBody: `# Hello`,
			wantSkip: 0,
		},
		{
			name: "incomplete frontmatter",
			input: `---
title: "Oops"`,
			wantBody: `---
title: "Oops"`,
			wantSkip: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, skip := StripFrontmatter(tt.input)
			if body != tt.wantBody {
				t.Errorf("got body %q, want %q", body, tt.wantBody)
			}
			if skip != tt.wantSkip {
				t.Errorf("got skip %d, want %d", skip, tt.wantSkip)
			}
		})
	}
}
