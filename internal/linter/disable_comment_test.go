package linter

import (
	"testing"
)

func TestParseDirectiveLine(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		wantDirective string
		wantRules     []string
	}{
		{
			name:          "no comment",
			line:          "plain text",
			wantDirective: "",
		},
		{
			name:          "unrelated comment",
			line:          "<!-- some comment -->",
			wantDirective: "",
		},
		{
			name:          "disable all",
			line:          "<!-- gomarklint-disable -->",
			wantDirective: "disable",
			wantRules:     []string{},
		},
		{
			name:          "disable named rules",
			line:          "<!-- gomarklint-disable no-bare-urls heading-level -->",
			wantDirective: "disable",
			wantRules:     []string{"no-bare-urls", "heading-level"},
		},
		{
			name:          "enable all",
			line:          "<!-- gomarklint-enable -->",
			wantDirective: "enable",
			wantRules:     []string{},
		},
		{
			name:          "enable named rule",
			line:          "<!-- gomarklint-enable no-bare-urls -->",
			wantDirective: "enable",
			wantRules:     []string{"no-bare-urls"},
		},
		{
			name:          "disable-line all",
			line:          "text <!-- gomarklint-disable-line -->",
			wantDirective: "disable-line",
			wantRules:     []string{},
		},
		{
			name:          "disable-line named rule",
			line:          "https://example.com <!-- gomarklint-disable-line no-bare-urls -->",
			wantDirective: "disable-line",
			wantRules:     []string{"no-bare-urls"},
		},
		{
			name:          "disable-next-line all",
			line:          "<!-- gomarklint-disable-next-line -->",
			wantDirective: "disable-next-line",
			wantRules:     []string{},
		},
		{
			name:          "disable-next-line named rule",
			line:          "<!-- gomarklint-disable-next-line duplicate-heading -->",
			wantDirective: "disable-next-line",
			wantRules:     []string{"duplicate-heading"},
		},
		{
			name:          "unknown directive",
			line:          "<!-- gomarklint-ignore -->",
			wantDirective: "",
		},
		{
			name:          "unclosed comment",
			line:          "<!-- gomarklint-disable",
			wantDirective: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDir, gotRules := parseDirectiveLine(tt.line)
			if gotDir != tt.wantDirective {
				t.Errorf("directive = %q, want %q", gotDir, tt.wantDirective)
			}
			if tt.wantDirective == "" {
				return
			}
			if len(gotRules) != len(tt.wantRules) {
				t.Fatalf("rules = %v, want %v", gotRules, tt.wantRules)
			}
			for i, r := range tt.wantRules {
				if gotRules[i] != r {
					t.Errorf("rules[%d] = %q, want %q", i, gotRules[i], r)
				}
			}
		})
	}
}

func TestParseDisableComments_BlockDisableAll(t *testing.T) {
	lines := []string{
		"# Heading",                   // line 1
		"<!-- gomarklint-disable -->", // line 2
		"https://example.com",         // line 3
		"<!-- gomarklint-enable -->",  // line 4
		"https://example.com",         // line 5
	}
	set := parseDisableComments(lines, 0)

	if set.isDisabled(1, "no-bare-urls") {
		t.Error("line 1 should not be disabled")
	}
	if !set.isDisabled(2, "no-bare-urls") {
		t.Error("line 2 (directive line) should be disabled")
	}
	if !set.isDisabled(3, "no-bare-urls") {
		t.Error("line 3 should be disabled")
	}
	if !set.isDisabled(3, "heading-level") {
		t.Error("line 3 should be disabled for all rules")
	}
	if set.isDisabled(4, "no-bare-urls") {
		t.Error("line 4 (enable directive) should not be disabled")
	}
	if set.isDisabled(5, "no-bare-urls") {
		t.Error("line 5 after enable should not be disabled")
	}
}

func TestParseDisableComments_BlockDisableNamedRule(t *testing.T) {
	lines := []string{
		"<!-- gomarklint-disable no-bare-urls -->", // line 1
		"https://example.com",                      // line 2
		"<!-- gomarklint-enable no-bare-urls -->",  // line 3
		"https://example.com",                      // line 4
	}
	set := parseDisableComments(lines, 0)

	if !set.isDisabled(2, "no-bare-urls") {
		t.Error("line 2 should be disabled for no-bare-urls")
	}
	if set.isDisabled(2, "heading-level") {
		t.Error("line 2 should not be disabled for heading-level")
	}
	if set.isDisabled(4, "no-bare-urls") {
		t.Error("line 4 after enable should not be disabled")
	}
}

func TestParseDisableComments_DisableLine(t *testing.T) {
	lines := []string{
		"# Heading", // line 1
		"https://example.com <!-- gomarklint-disable-line -->", // line 2
		"https://example.com", // line 3
	}
	set := parseDisableComments(lines, 0)

	if set.isDisabled(1, "no-bare-urls") {
		t.Error("line 1 should not be disabled")
	}
	if !set.isDisabled(2, "no-bare-urls") {
		t.Error("line 2 should be disabled")
	}
	if !set.isDisabled(2, "heading-level") {
		t.Error("line 2 should be disabled for all rules")
	}
	if set.isDisabled(3, "no-bare-urls") {
		t.Error("line 3 should not be disabled")
	}
}

func TestParseDisableComments_DisableNextLine(t *testing.T) {
	lines := []string{
		"# Heading",                             // line 1
		"<!-- gomarklint-disable-next-line -->", // line 2
		"https://example.com",                   // line 3
		"https://example.com",                   // line 4
	}
	set := parseDisableComments(lines, 0)

	if set.isDisabled(1, "no-bare-urls") {
		t.Error("line 1 should not be disabled")
	}
	if set.isDisabled(2, "no-bare-urls") {
		t.Error("line 2 (directive itself) should not be disabled")
	}
	if !set.isDisabled(3, "no-bare-urls") {
		t.Error("line 3 should be disabled")
	}
	if set.isDisabled(4, "no-bare-urls") {
		t.Error("line 4 should not be disabled")
	}
}

func TestParseDisableComments_WithFrontmatterOffset(t *testing.T) {
	lines := []string{
		"<!-- gomarklint-disable-next-line no-bare-urls -->", // body line 1 → abs line 4
		"https://example.com",                                // body line 2 → abs line 5
	}
	offset := 3
	set := parseDisableComments(lines, offset)

	if set.isDisabled(4, "no-bare-urls") {
		t.Error("abs line 4 (directive) should not be disabled")
	}
	if !set.isDisabled(5, "no-bare-urls") {
		t.Error("abs line 5 should be disabled")
	}
}

func TestDisabledSet_IsDisabled(t *testing.T) {
	set := make(disabledSet)
	set[10] = nil // all rules

	if !set.isDisabled(10, "any-rule") {
		t.Error("nil entry should disable all rules")
	}
	if set.isDisabled(11, "any-rule") {
		t.Error("unregistered line should not be disabled")
	}

	set[20] = map[string]struct{}{"no-bare-urls": {}}
	if !set.isDisabled(20, "no-bare-urls") {
		t.Error("named rule should be disabled")
	}
	if set.isDisabled(20, "heading-level") {
		t.Error("other rule should not be disabled")
	}
}
