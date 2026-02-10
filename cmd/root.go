package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/internal/config"
	"github.com/shinagawa-web/gomarklint/internal/file"
	"github.com/shinagawa-web/gomarklint/internal/linter"
	"github.com/shinagawa-web/gomarklint/internal/output"
)

var minHeadingLevel int
var enableLinkCheck bool
var skipLinkPatterns []string
var configFilePath string
var outputFormat string
var enableHeadingLevelCheck bool
var enableDuplicateHeadingCheck bool
var enableNoMultipleBlankLinesCheck bool
var enableNoSetextHeadingsCheck bool
var enableFinalBlankLineCheck bool

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
		MinHeadingLevel:                 minHeadingLevel,
		EnableLinkCheck:                 enableLinkCheck,
		EnableHeadingLevelCheck:         enableHeadingLevelCheck,
		EnableDuplicateHeadingCheck:     enableDuplicateHeadingCheck,
		EnableNoMultipleBlankLinesCheck: enableNoMultipleBlankLinesCheck,
		EnableNoSetextHeadingsCheck:     enableNoSetextHeadingsCheck,
		EnableFinalBlankLineCheck:       enableFinalBlankLineCheck,
		SkipLinkPatterns:                skipLinkPatterns,
		OutputFormat:                    outputFormat,
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
		return errors.New("")
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
	if !cfg.EnableLinkCheck {
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

	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&enableLinkCheck, "enable-link-check", false, "enable external link checking")
	rootCmd.Flags().BoolVar(&enableHeadingLevelCheck, "enable-heading-level-check", true, "enable heading level check")
	rootCmd.Flags().BoolVar(&enableDuplicateHeadingCheck, "enable-duplicate-heading-check", true, "enable duplicate heading check")
	rootCmd.Flags().BoolVar(&enableNoMultipleBlankLinesCheck, "enable-no-multiple-blank-lines-check", true, "enable no multiple blank lines check")
	rootCmd.Flags().BoolVar(&enableNoSetextHeadingsCheck, "enable-no-setext-headings-check", true, "enable no setext headings check")
	rootCmd.Flags().BoolVar(&enableFinalBlankLineCheck, "enable-final-blank-line-check", true, "enable final blank line check")
	rootCmd.Flags().StringArrayVar(&skipLinkPatterns, "skip-link-patterns", nil, "patterns of URLs to skip link checking")
	rootCmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text or json")

	rootCmd.AddCommand(initCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
