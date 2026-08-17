package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type FlagValues struct {
	OutputFormat string
	MinSeverity  string
}

func LoadOrDefault(configPath string) (Config, error) {
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("failed to access config file: %w", err)
	}
	return LoadConfig(configPath)
}

func MergeFlags(cfg Config, cmd *cobra.Command, flags FlagValues) Config {
	if cmd.Flags().Changed("output") {
		cfg.OutputFormat = flags.OutputFormat
	}
	if cmd.Flags().Changed("severity") {
		cfg.MinSeverity = RuleSeverity(flags.MinSeverity)
	}
	return cfg
}

func Validate(cfg Config) error {
	if cfg.OutputFormat != "text" && cfg.OutputFormat != "json" {
		return fmt.Errorf("invalid output format: %q (must be 'text' or 'json')", cfg.OutputFormat)
	}
	switch cfg.MinSeverity {
	case SeverityWarning, SeverityError:
	default:
		return fmt.Errorf("invalid severity: %q (must be 'warning' or 'error')", cfg.MinSeverity)
	}
	return nil
}
