package config

import (
	"encoding/json"
	"fmt"
)

// DefaultConfigJSON is the canonical JSON written by `gomarklint init`.
// It must stay in sync with Default() — update both together when adding rules.
const DefaultConfigJSON = `{
  "default": true,
  "rules": {
    "final-blank-line": true,
    "unclosed-code-block": true,
    "empty-alt-text": true,
    "fenced-code-language": true,
    "heading-level": { "severity": "error", "minLevel": 2 },
    "duplicate-heading": true,
    "no-multiple-blank-lines": true,
    "no-setext-headings": true,
    "external-link": { "enabled": false, "severity": "error", "timeoutSeconds": 5, "skipPatterns": [] }
  },
  "include": ["README.md", "testdata"],
  "ignore": [],
  "output": "text"
}
`

// RuleSeverity represents the severity level of a rule violation.
type RuleSeverity string

const (
	SeverityError   RuleSeverity = "error"
	SeverityWarning RuleSeverity = "warning"
	SeverityOff     RuleSeverity = "off"
)

// RuleConfig holds per-rule configuration.
// It supports three JSON shorthand forms:
//
//	true             → enabled, severity = "error"
//	false            → disabled
//	"warning"        → enabled, severity = "warning"
//	{"enabled": true, "severity": "warning", ...options}
type RuleConfig struct {
	Enabled  bool
	Severity RuleSeverity
	Options  map[string]interface{}
}

// UnmarshalJSON handles bool, string, and object forms.
func (r *RuleConfig) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		return r.fromBool(b)
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return r.fromString(s)
	}
	return r.fromObject(data)
}

func (r *RuleConfig) fromBool(b bool) error {
	r.Enabled = b
	if b {
		r.Severity = SeverityError
	} else {
		r.Severity = SeverityOff
	}
	return nil
}

func (r *RuleConfig) fromString(s string) error {
	switch RuleSeverity(s) {
	case SeverityError:
		r.Enabled = true
		r.Severity = SeverityError
	case SeverityWarning:
		r.Enabled = true
		r.Severity = SeverityWarning
	case SeverityOff:
		r.Enabled = false
		r.Severity = SeverityOff
	default:
		return fmt.Errorf("invalid rule value: %q (use true, false, \"error\", \"warning\", or \"off\")", s)
	}
	return nil
}

func (r *RuleConfig) fromObject(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid rule config: %w", err)
	}
	r.Enabled = true
	r.Severity = SeverityError
	r.Options = map[string]interface{}{}
	for k, v := range raw {
		if err := r.applyObjectField(k, v); err != nil {
			return err
		}
	}
	// enabled=false always forces SeverityOff; severity=off always forces Enabled=false
	if !r.Enabled {
		r.Severity = SeverityOff
	}
	if r.Severity == SeverityOff {
		r.Enabled = false
	}
	return nil
}

func (r *RuleConfig) applyObjectField(k string, v json.RawMessage) error {
	switch k {
	case "enabled":
		var b bool
		if err := json.Unmarshal(v, &b); err != nil {
			return fmt.Errorf("invalid \"enabled\" value: %w", err)
		}
		r.Enabled = b
	case "severity":
		var sev string
		if err := json.Unmarshal(v, &sev); err != nil {
			return fmt.Errorf("invalid \"severity\" value: %w", err)
		}
		switch RuleSeverity(sev) {
		case SeverityError, SeverityWarning, SeverityOff:
			r.Severity = RuleSeverity(sev)
		default:
			return fmt.Errorf("invalid severity: %q (use \"error\", \"warning\", or \"off\")", sev)
		}
	default:
		var val interface{}
		_ = json.Unmarshal(v, &val)
		r.Options[k] = val
	}
	return nil
}

// Config defines the options for gomarklint, loaded from a config file.
type Config struct {
	// Default controls whether rules are enabled by default when not listed in Rules.
	// true = all rules on by default; false = opt-in mode (only listed rules run).
	Default bool `json:"default"`

	// Rules maps rule keys to their configuration.
	Rules map[string]*RuleConfig `json:"rules"`

	// Include lists files or directories to lint when no arguments are given.
	Include []string `json:"include"`

	// Ignore lists glob patterns to exclude from linting.
	Ignore []string `json:"ignore"`

	// OutputFormat controls output: "text" or "json".
	OutputFormat string `json:"output"`

	// MinSeverity filters output: only report rules at or above this severity ("warning" or "error").
	// This field is not serialized to JSON; it is set via CLI flag only.
	MinSeverity RuleSeverity `json:"-"`
}

// IsEnabled reports whether the named rule should run.
func (c *Config) IsEnabled(name string) bool {
	rc, ok := c.Rules[name]
	if !ok || rc == nil {
		return c.Default
	}
	return rc.Enabled
}

// RuleOptions returns the options map for the named rule, or an empty map.
func (c *Config) RuleOptions(name string) map[string]interface{} {
	rc, ok := c.Rules[name]
	if !ok || rc == nil || rc.Options == nil {
		return map[string]interface{}{}
	}
	return rc.Options
}

// RuleSeverity returns the configured severity for the named rule.
// Returns "error" if the rule is not listed or has no severity set.
func (c *Config) RuleSeverity(name string) string {
	rc, ok := c.Rules[name]
	if !ok || rc == nil || rc.Severity == "" {
		return string(SeverityError)
	}
	return string(rc.Severity)
}

// Default returns the default configuration with all standard rules enabled.
func Default() Config {
	return Config{
		Default: true,
		Rules: map[string]*RuleConfig{
			"final-blank-line": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"unclosed-code-block": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"empty-alt-text": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"fenced-code-language": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"heading-level": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"minLevel": float64(2)},
			},
			"duplicate-heading": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"no-multiple-blank-lines": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"no-setext-headings": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{},
			},
			"external-link": {
				Enabled:  false,
				Severity: SeverityError,
				Options:  map[string]interface{}{"timeoutSeconds": float64(5), "skipPatterns": []interface{}{}},
			},
		},
		Include:      []string{"README.md", "testdata"},
		Ignore:       []string{},
		OutputFormat: "text",
		MinSeverity:  SeverityWarning,
	}
}
