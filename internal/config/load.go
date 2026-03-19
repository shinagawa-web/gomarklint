package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig reads the config file from the given path and returns a Config object.
// If the file is missing or unreadable, an error is returned.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	cfg := Config{Default: true}
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "text"
	}
	if cfg.MinSeverity == "" {
		cfg.MinSeverity = SeverityWarning
	}
	if cfg.Rules == nil {
		// rules key was omitted entirely — seed from built-in defaults so that
		// rules like external-link remain disabled by default.
		cfg.Rules = Default().Rules
	}

	return cfg, nil
}
