package config

import (
	"encoding/json"
	"fmt"
)

const DefaultNoTrailingPunctuation = ".,;:!"

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
    "single-h1": true,
    "blanks-around-headings": true,
    "no-bare-urls": true,
    "no-empty-links": true,
    "no-emphasis-as-heading": true,
    "blanks-around-lists": true,
    "blanks-around-fences": true,
    "no-hard-tabs": true,
    "no-trailing-punctuation": { "punctuation": ".,;:!" },
    "consistent-code-fence": { "style": "consistent" },
    "consistent-emphasis-style": { "style": "consistent" },
    "consistent-list-marker": { "style": "consistent" },
    "max-line-length": { "enabled": false, "lineLength": 80 },
    "external-link": { "enabled": false, "severity": "error", "timeoutSeconds": 5, "maxConcurrency": 10, "maxRetries": 2, "perHostConcurrency": 2, "perHostIntervalMs": 3000, "skipPatterns": [] },
    "link-fragments": { "enabled": true, "slug-algorithm": "github" }
  },
  "include": ["README.md", "testdata"],
  "ignore": [],
  "output": "text"
}
`

type RuleSeverity string

const (
	SeverityError   RuleSeverity = "error"
	SeverityWarning RuleSeverity = "warning"
	SeverityOff     RuleSeverity = "off"
)

type RuleConfig struct {
	Enabled  bool
	Severity RuleSeverity
	Options  map[string]interface{}
}

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

type Config struct {
	Default      bool                   `json:"default"`
	Rules        map[string]*RuleConfig `json:"rules"`
	Include      []string               `json:"include"`
	Ignore       []string               `json:"ignore"`
	OutputFormat string                 `json:"output"`
	MinSeverity  RuleSeverity           `json:"-"`
}

func (c *Config) IsEnabled(name string) bool {
	rc, ok := c.Rules[name]
	if !ok || rc == nil {
		return c.Default
	}
	return rc.Enabled
}

func (c *Config) RuleOptions(name string) map[string]interface{} {
	rc, ok := c.Rules[name]
	if !ok || rc == nil || rc.Options == nil {
		return map[string]interface{}{}
	}
	return rc.Options
}

func (c *Config) RuleSeverity(name string) string {
	rc, ok := c.Rules[name]
	if !ok || rc == nil || rc.Severity == "" {
		return string(SeverityError)
	}
	return string(rc.Severity)
}

func enabledRule() *RuleConfig {
	return &RuleConfig{Enabled: true, Severity: SeverityError, Options: map[string]interface{}{}}
}

func Default() Config {
	return Config{
		Default: true,
		Rules: map[string]*RuleConfig{
			"final-blank-line":        enabledRule(),
			"unclosed-code-block":     enabledRule(),
			"empty-alt-text":          enabledRule(),
			"fenced-code-language":    enabledRule(),
			"duplicate-heading":       enabledRule(),
			"no-multiple-blank-lines": enabledRule(),
			"no-setext-headings":      enabledRule(),
			"single-h1":               enabledRule(),
			"blanks-around-headings":  enabledRule(),
			"no-bare-urls":            enabledRule(),
			"no-empty-links":          enabledRule(),
			"no-emphasis-as-heading":  enabledRule(),
			"blanks-around-lists":     enabledRule(),
			"blanks-around-fences":    enabledRule(),
			"no-hard-tabs":            enabledRule(),
			"no-trailing-punctuation": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"punctuation": DefaultNoTrailingPunctuation},
			},
			"consistent-code-fence": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"style": "consistent"},
			},
			"consistent-emphasis-style": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"style": "consistent"},
			},
			"consistent-list-marker": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"style": "consistent"},
			},
			"max-line-length": {
				Enabled:  false,
				Severity: SeverityOff,
				Options:  map[string]interface{}{"lineLength": float64(80)},
			},
			"heading-level": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"minLevel": float64(2)},
			},
			"external-link": {
				Enabled:  false,
				Severity: SeverityError,
				Options:  map[string]interface{}{"timeoutSeconds": float64(5), "maxConcurrency": float64(10), "maxRetries": float64(2), "perHostConcurrency": float64(2), "perHostIntervalMs": float64(3000), "skipPatterns": []interface{}{}},
			},
			"link-fragments": {
				Enabled:  true,
				Severity: SeverityError,
				Options:  map[string]interface{}{"slug-algorithm": "github"},
			},
		},
		Include:      []string{"README.md", "testdata"},
		Ignore:       []string{},
		OutputFormat: "text",
		MinSeverity:  SeverityWarning,
	}
}
