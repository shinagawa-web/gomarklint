package cmd

import (
	"fmt"
	"os"

	"github.com/shinagawa-web/gomarklint/internal/parser"
	"github.com/shinagawa-web/gomarklint/internal/rule"
	"github.com/spf13/cobra"
)

var minHeadingLevel int
var checkLinks bool

var rootCmd = &cobra.Command{
	Use:   "gomarklint",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a markdown file or directory")
		}
		files, err := parser.ExpandPaths(args)
		if err != nil {
			return fmt.Errorf("failed to expand paths: %w", err)
		}
		for _, path := range files {
			fmt.Printf("Linting: %s\n", path)
			content, err := parser.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
				continue
			}
			allErrors := []rule.LintError{}
			allErrors = append(allErrors, rule.CheckHeadingLevels(path, content, minHeadingLevel)...)
			allErrors = append(allErrors, rule.CheckFinalBlankLine(path, content)...)
			allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, content)...)
			if checkLinks {
				allErrors = append(allErrors, rule.CheckExternalLinks(path, content)...)
			}
			if len(allErrors) == 0 {
				fmt.Println("No issues found 🎉")
			} else {
				for _, e := range allErrors {
					fmt.Printf("%s:%d: %s\n", e.File, e.Line, e.Message)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().IntVar(&minHeadingLevel, "min-heading", 2, "minimum heading level to start from (default: 2)")
	rootCmd.Flags().BoolVar(&checkLinks, "check-links", false, "enable external link checking")
}

func Execute() error {
	return rootCmd.Execute()
}
