package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// FlagValues holds the command-line flag values.
type FlagValues struct {
	OutputFormat string
	MinSeverity  string
}

// LoadOrDefault loads configuration from file if it exists, otherwise returns default config.
func LoadOrDefault(configPath string) (Config, error) {
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("failed to access config file: %w", err)
	}
	return LoadConfig(configPath)
}

// MergeFlags merges command-line flag values into the config, respecting which flags were actually set.
func MergeFlags(cfg Config, cmd *cobra.Command, flags FlagValues) Config {
	if cmd.Flags().Changed("output") {
		cfg.OutputFormat = flags.OutputFormat
	}
	if cmd.Flags().Changed("severity") {
		cfg.MinSeverity = RuleSeverity(flags.MinSeverity)
	}
	return cfg
}

// Validate checks if the configuration values are valid.
func Validate(cfg Config) error {
	if cfg.OutputFormat != "text" && cfg.OutputFormat != "json" {
		return fmt.Errorf("invalid output format: %q (must be 'text' or 'json')", cfg.OutputFormat)
	}
	switch cfg.MinSeverity {
	case SeverityWarning, SeverityError:
		// valid
	default:
		return fmt.Errorf("invalid severity: %q (must be 'warning' or 'error')", cfg.MinSeverity)
	}
	return nil
}
