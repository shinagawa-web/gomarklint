package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
	"github.com/shinagawa-web/gomarklint/v2/internal/file"
	"github.com/shinagawa-web/gomarklint/v2/internal/linter"
	"github.com/shinagawa-web/gomarklint/v2/internal/output"
)

// ErrLintViolations is returned when lint violations are found.
var ErrLintViolations = errors.New("lint violations found")

var configFilePath string
var outputFormat string
var minSeverity string

var rootCmd = &cobra.Command{
	Use:   "gomarklint [files or directories]",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runLint,
}

func runLint(cmd *cobra.Command, args []string) error {
	fmt.Println()
	start := time.Now()

	// Load and merge configuration
	cfg, err := config.LoadOrDefault(configFilePath)
	if err != nil {
		return err
	}

	flags := config.FlagValues{
		OutputFormat: outputFormat,
		MinSeverity:  minSeverity,
	}
	cfg = config.MergeFlags(cfg, cmd, flags)

	if err := config.Validate(cfg); err != nil {
		return err
	}

	// Determine files to check
	if len(args) == 0 {
		if len(cfg.Include) > 0 {
			args = cfg.Include
		} else {
			return fmt.Errorf("please provide a markdown file or directory (or set 'include' in .gomarklint.json)")
		}
	}

	files, err := file.ExpandPaths(args, cfg.Ignore)
	if err != nil {
		return fmt.Errorf("failed to expand paths: %w", err)
	}

	// Run linter
	lint, err := linter.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create linter: %w", err)
	}

	result, err := lint.Run(files)
	if err != nil {
		return err
	}

	// Format and output results
	if err := formatOutput(cfg, result, len(files), time.Since(start)); err != nil {
		return err
	}

	if result.TotalErrors > 0 {
		return ErrLintViolations
	}
	return nil
}

func formatOutput(cfg config.Config, result *linter.Result, fileCount int, duration time.Duration) error {
	var formatter output.Formatter
	if cfg.OutputFormat == "json" {
		formatter = output.NewJSONFormatter()
	} else {
		formatter = output.NewTextFormatter()
	}

	linksChecked := &result.TotalLinksChecked
	if !cfg.IsEnabled("external-link") {
		linksChecked = nil
	}

	outputResult := &output.Result{
		Files:        fileCount,
		Lines:        result.TotalLines,
		Errors:       result.TotalErrors,
		LinksChecked: linksChecked,
		Duration:     duration,
		Details:      result.Errors,
		OrderedPaths: result.OrderedPaths,
	}

	return formatter.Format(os.Stdout, outputResult)
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.Flags().StringVar(&configFilePath, "config", ".gomarklint.json", "path to config file (default: .gomarklint.json)")
	rootCmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text or json")
	rootCmd.Flags().StringVar(&minSeverity, "severity", "warning", "minimum severity to report: warning or error")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
