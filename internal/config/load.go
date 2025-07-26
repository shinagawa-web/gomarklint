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

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}
	if cfg.OutputFormat == "" {
		// Fallback to default if not set in config file
		cfg.OutputFormat = "text"
	}

	return cfg, nil
}
