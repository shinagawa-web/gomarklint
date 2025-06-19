package cmd

import (
	"fmt"

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
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
