package cmd

import (
	"fmt"
	"os"

	"github.com/shinagawa-web/gomarklint/internal/parser"
	"github.com/shinagawa-web/gomarklint/internal/rule"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gomarklint",
	Short: "A fast markdown linter written in Go",
	Long:  "gomarklint checks markdown files for common issues like heading structure, blank lines, and more.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a markdown file or directory")
		}
		for _, path := range args {
			fmt.Printf("Linting: %s\n", path)
			content, err := parser.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
				continue
			}

			errors := rule.CheckHeadingLevels(content)
			if len(errors) == 0 {
				fmt.Println("No issues found 🎉")
			} else {
				for _, e := range errors {
					fmt.Printf("%s:%d: %s\n", path, e.Line, e.Message)
				}
			}
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
