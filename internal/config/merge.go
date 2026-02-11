package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// FlagValues holds the command-line flag values.
type FlagValues struct {
	MinHeadingLevel                 int
	EnableLinkCheck                 bool
	EnableHeadingLevelCheck         bool
	EnableDuplicateHeadingCheck     bool
	EnableNoMultipleBlankLinesCheck bool
	EnableNoSetextHeadingsCheck     bool
	EnableFinalBlankLineCheck       bool
	SkipLinkPatterns                []string
	OutputFormat                    string
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
	if cmd.Flags().Changed("min-heading") {
		cfg.MinHeadingLevel = flags.MinHeadingLevel
	}
	if cmd.Flags().Changed("enable-link-check") {
		cfg.EnableLinkCheck = flags.EnableLinkCheck
	}
	if cmd.Flags().Changed("enable-heading-level-check") {
		cfg.EnableHeadingLevelCheck = flags.EnableHeadingLevelCheck
	}
	if cmd.Flags().Changed("enable-duplicate-heading-check") {
		cfg.EnableDuplicateHeadingCheck = flags.EnableDuplicateHeadingCheck
	}
	if cmd.Flags().Changed("enable-no-multiple-blank-lines-check") {
		cfg.EnableNoMultipleBlankLinesCheck = flags.EnableNoMultipleBlankLinesCheck
	}
	if cmd.Flags().Changed("enable-no-setext-headings-check") {
		cfg.EnableNoSetextHeadingsCheck = flags.EnableNoSetextHeadingsCheck
	}
	if cmd.Flags().Changed("enable-final-blank-line-check") {
		cfg.EnableFinalBlankLineCheck = flags.EnableFinalBlankLineCheck
	}
	if cmd.Flags().Changed("skip-link-patterns") {
		cfg.SkipLinkPatterns = flags.SkipLinkPatterns
	}
	if cmd.Flags().Changed("output") {
		cfg.OutputFormat = flags.OutputFormat
	}
	return cfg
}

// Validate checks if the configuration values are valid.
func Validate(cfg Config) error {
	if cfg.OutputFormat != "text" && cfg.OutputFormat != "json" {
		return fmt.Errorf("invalid output format: %q (must be 'text' or 'json')", cfg.OutputFormat)
	}
	return nil
}
