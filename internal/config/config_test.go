package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	json := `{
		"default": true,
		"rules": {
			"heading-level": { "enabled": true, "severity": "error", "minLevel": 3 },
			"duplicate-heading": true,
			"external-link": { "enabled": true, "timeoutSeconds": 10 }
		},
		"output": "text"
	}`

	tmp := filepath.Join(t.TempDir(), ".gomarklint.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// heading-level: full object form
	hl := cfg.Rules["heading-level"]
	if hl == nil {
		t.Fatal("expected heading-level rule")
		return
	}
	if !hl.Enabled {
		t.Error("expected heading-level enabled=true")
	}
	if hl.Severity != SeverityError {
		t.Errorf("expected heading-level severity=error, got %s", hl.Severity)
	}
	if v, ok := hl.Options["minLevel"]; !ok || v.(float64) != 3 {
		t.Errorf("expected heading-level minLevel=3, got %v", v)
	}

	// duplicate-heading: bool shorthand
	dh := cfg.Rules["duplicate-heading"]
	if dh == nil || !dh.Enabled {
		t.Error("expected duplicate-heading enabled=true")
	}
	if dh.Severity != SeverityError {
		t.Errorf("expected duplicate-heading severity=error, got %s", dh.Severity)
	}

	// external-link: object with options
	el := cfg.Rules["external-link"]
	if el == nil || !el.Enabled {
		t.Error("expected external-link enabled=true")
	}
	if v, ok := el.Options["timeoutSeconds"]; !ok || v.(float64) != 10 {
		t.Errorf("expected external-link timeoutSeconds=10, got %v", v)
	}

	// OutputFormat defaults to "text"
	if cfg.OutputFormat != "text" {
		t.Errorf("expected OutputFormat=text, got %s", cfg.OutputFormat)
	}
}

func TestLoadConfig_RuleShorthands(t *testing.T) {
	json := `{
		"default": false,
		"rules": {
			"final-blank-line": true,
			"unclosed-code-block": false,
			"no-setext-headings": "warning",
			"no-multiple-blank-lines": "error",
			"empty-alt-text": "off"
		}
	}`

	tmp := filepath.Join(t.TempDir(), ".gomarklint.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// bool true
	if r := cfg.Rules["final-blank-line"]; r == nil || !r.Enabled || r.Severity != SeverityError {
		t.Error("expected final-blank-line: true → enabled, error")
	}
	// bool false
	if r := cfg.Rules["unclosed-code-block"]; r == nil || r.Enabled || r.Severity != SeverityOff {
		t.Error("expected unclosed-code-block: false → disabled, off")
	}
	// string "warning"
	if r := cfg.Rules["no-setext-headings"]; r == nil || !r.Enabled || r.Severity != SeverityWarning {
		t.Error("expected no-setext-headings: \"warning\" → enabled, warning")
	}
	// string "error"
	if r := cfg.Rules["no-multiple-blank-lines"]; r == nil || !r.Enabled || r.Severity != SeverityError {
		t.Error("expected no-multiple-blank-lines: \"error\" → enabled, error")
	}
	// string "off"
	if r := cfg.Rules["empty-alt-text"]; r == nil || r.Enabled || r.Severity != SeverityOff {
		t.Error("expected empty-alt-text: \"off\" → disabled, off")
	}

	// default=false: unlisted rule should be disabled
	if cfg.IsEnabled("heading-level") {
		t.Error("expected heading-level to be disabled (default=false, not listed)")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmp, []byte(`{ invalid json }`), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(tmp)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadConfig_InvalidRuleValue(t *testing.T) {
	json := `{"default": true, "rules": {"heading-level": "unknown"}}`
	tmp := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(tmp)
	if err == nil {
		t.Error("expected error for invalid rule value, got nil")
	}
}

func TestLoadConfig_UnknownTopLevelField(t *testing.T) {
	json := `{"unknown": true}`
	tmp := filepath.Join(t.TempDir(), "unknown.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(tmp)
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if !cfg.Default {
		t.Error("expected Default=true")
	}
	if cfg.OutputFormat != "text" {
		t.Errorf("expected OutputFormat=text, got %s", cfg.OutputFormat)
	}
	if cfg.MinSeverity != SeverityWarning {
		t.Errorf("expected MinSeverity=warning, got %s", cfg.MinSeverity)
	}

	for _, name := range []string{
		"final-blank-line", "unclosed-code-block", "empty-alt-text",
		"heading-level", "duplicate-heading", "no-multiple-blank-lines",
		"no-setext-headings", "external-link",
	} {
		if _, ok := cfg.Rules[name]; !ok {
			t.Errorf("expected rule %q in Default()", name)
		}
	}

	// external-link off by default
	if cfg.IsEnabled("external-link") {
		t.Error("expected external-link to be disabled by default")
	}

	// heading-level minLevel=2
	if v, ok := cfg.RuleOptions("heading-level")["minLevel"]; !ok || v.(float64) != 2 {
		t.Errorf("expected heading-level minLevel=2, got %v", v)
	}
}

func TestLoadConfig_NoRulesKey(t *testing.T) {
	// When "rules" is omitted, cfg.Rules is seeded from Default().Rules (not nil).
	// This ensures opt-out rules like external-link remain disabled even when
	// default=true would otherwise enable all unlisted rules via IsEnabled fallback.
	json := `{"default": true, "output": "json"}`
	tmp := filepath.Join(t.TempDir(), "norules.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules == nil {
		t.Error("expected Rules to be seeded from Default().Rules, got nil")
	}
	// external-link must remain disabled (opt-in rule)
	if cfg.IsEnabled("external-link") {
		t.Error("expected external-link to be disabled even when default=true and rules key is omitted")
	}
}

func TestLoadConfig_PartialRulesKey_ExternalLinkDisabledByDefault(t *testing.T) {
	// When "rules" is provided but external-link is not listed, it should remain
	// disabled by default (opt-in rule), not enabled by the default=true fallback.
	json := `{"default": true, "rules": {"heading-level": true}}`
	tmp := filepath.Join(t.TempDir(), "partial.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IsEnabled("external-link") {
		t.Error("expected external-link to be disabled when not listed in rules, even with default=true")
	}
	// explicitly listed rule should still work
	if !cfg.IsEnabled("heading-level") {
		t.Error("expected heading-level to be enabled")
	}
}

func TestUnmarshalJSON_ObjectForm(t *testing.T) {
	t.Run("EnabledFalse sets SeverityOff", func(t *testing.T) {
		json := `{"default": true, "rules": {"heading-level": {"enabled": false, "severity": "error"}}}`
		tmp := filepath.Join(t.TempDir(), "cfg.json")
		if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r := cfg.Rules["heading-level"]
		if r.Enabled {
			t.Error("expected Enabled=false")
		}
		if r.Severity != SeverityOff {
			t.Errorf("expected Severity=off when enabled=false, got %s", r.Severity)
		}
	})

	t.Run("InvalidEnabledValue", func(t *testing.T) {
		json := `{"default": true, "rules": {"heading-level": {"enabled": "yes"}}}`
		tmp := filepath.Join(t.TempDir(), "cfg.json")
		if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(tmp)
		if err == nil {
			t.Error("expected error for invalid enabled value")
		}
	})

	t.Run("InvalidSeverityValue", func(t *testing.T) {
		json := `{"default": true, "rules": {"heading-level": {"severity": "critical"}}}`
		tmp := filepath.Join(t.TempDir(), "cfg.json")
		if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(tmp)
		if err == nil {
			t.Error("expected error for invalid severity value")
		}
	})

	t.Run("InvalidSeverityType", func(t *testing.T) {
		json := `{"default": true, "rules": {"heading-level": {"severity": 42}}}`
		tmp := filepath.Join(t.TempDir(), "cfg.json")
		if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(tmp)
		if err == nil {
			t.Error("expected error for non-string severity type")
		}
	})

	t.Run("InvalidOptionValue", func(t *testing.T) {
		// A JSON object value that can't be unmarshaled (e.g., a function) can't
		// be triggered from JSON alone — but we can hit the options path with a
		// valid but unusual type like a nested object.
		json := `{"default": true, "rules": {"heading-level": {"minLevel": {"nested": true}}}}`
		tmp := filepath.Join(t.TempDir(), "cfg.json")
		if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Nested objects are valid interface{} values — just stored as-is.
		if cfg.Rules["heading-level"].Options["minLevel"] == nil {
			t.Error("expected minLevel option to be stored")
		}
	})

	t.Run("InvalidObjectJSON", func(t *testing.T) {
		var r RuleConfig
		err := r.UnmarshalJSON([]byte(`{invalid`))
		if err == nil {
			t.Error("expected error for malformed object JSON")
		}
	})
}

func TestRuleOptions_NilOptions(t *testing.T) {
	cfg := Config{
		Default: true,
		Rules: map[string]*RuleConfig{
			"heading-level": {Enabled: true, Severity: SeverityError, Options: nil},
		},
	}
	// Should return empty map, not panic
	opts := cfg.RuleOptions("heading-level")
	if opts == nil {
		t.Error("expected non-nil map for nil Options")
	}
	if len(opts) != 0 {
		t.Errorf("expected empty map, got %v", opts)
	}
}

func TestIsEnabled(t *testing.T) {
	cfg := Config{
		Default: true,
		Rules: map[string]*RuleConfig{
			"off-rule": {Enabled: false, Severity: SeverityOff},
		},
	}

	if cfg.IsEnabled("off-rule") {
		t.Error("expected off-rule to be disabled")
	}
	if !cfg.IsEnabled("some-other-rule") {
		t.Error("expected unlisted rule to be enabled (Default=true)")
	}

	cfg.Default = false
	if cfg.IsEnabled("some-other-rule") {
		t.Error("expected unlisted rule to be disabled (Default=false)")
	}
}

func TestIsEnabled_NilRuleEntry(t *testing.T) {
	cfg := Config{
		Default: true,
		Rules:   map[string]*RuleConfig{"some-rule": nil},
	}
	// nil entry should fall back to Default, not panic
	if !cfg.IsEnabled("some-rule") {
		t.Error("expected nil rule entry to fall back to Default=true")
	}
}

func TestRuleOptions_NilRuleEntry(t *testing.T) {
	cfg := Config{
		Default: true,
		Rules:   map[string]*RuleConfig{"some-rule": nil},
	}
	// nil entry should return empty map, not panic
	opts := cfg.RuleOptions("some-rule")
	if opts == nil {
		t.Error("expected non-nil map for nil rule entry")
	}
}

func TestRuleSeverity(t *testing.T) {
	cfg := Config{
		Default: true,
		Rules: map[string]*RuleConfig{
			"error-rule":   {Enabled: true, Severity: SeverityError},
			"warning-rule": {Enabled: true, Severity: SeverityWarning},
			"nil-rule":     nil,
		},
	}

	if s := cfg.RuleSeverity("error-rule"); s != "error" {
		t.Errorf("expected error, got %s", s)
	}
	if s := cfg.RuleSeverity("warning-rule"); s != "warning" {
		t.Errorf("expected warning, got %s", s)
	}
	// unlisted rule defaults to error
	if s := cfg.RuleSeverity("unlisted"); s != "error" {
		t.Errorf("expected error for unlisted rule, got %s", s)
	}
	// nil entry defaults to error
	if s := cfg.RuleSeverity("nil-rule"); s != "error" {
		t.Errorf("expected error for nil rule entry, got %s", s)
	}
}

func TestLoadConfig_DefaultKeyOmitted(t *testing.T) {
	// When "default" key is omitted, Default should be true (opt-out by default).
	json := `{"output": "text"}`
	tmp := filepath.Join(t.TempDir(), "nodefault.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Default {
		t.Error("expected Default=true when 'default' key is omitted")
	}
}

func TestUnmarshalJSON_SeverityOffDisablesRule(t *testing.T) {
	// severity="off" in object form should set Enabled=false
	json := `{"default": true, "rules": {"heading-level": {"severity": "off"}}}`
	tmp := filepath.Join(t.TempDir(), "cfg.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := cfg.Rules["heading-level"]
	if r.Enabled {
		t.Error("expected Enabled=false when severity=off")
	}
	if r.Severity != SeverityOff {
		t.Errorf("expected Severity=off, got %s", r.Severity)
	}
}
