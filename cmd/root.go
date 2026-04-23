package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/v2/internal/app"
	"github.com/shinagawa-web/gomarklint/v2/internal/config"
)

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
	opts := app.Options{
		ConfigPath: configFilePath,
		Args:       args,
	}
	if cmd.Flags().Changed("output") {
		opts.OutputFormat = outputFormat
	}
	if cmd.Flags().Changed("severity") {
		opts.MinSeverity = config.RuleSeverity(minSeverity)
	}
	return app.Run(os.Stdout, opts)
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("gomarklint version {{.Version}}\n")
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
