package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shinagawa-web/gomarklint/v2/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default .gomarklint.json config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		path := ".gomarklint.json"

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		if err := os.WriteFile(path, []byte(config.DefaultConfigJSON), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		green := "\033[32m"
		reset := "\033[0m"

		fmt.Printf("%s✔%s .gomarklint.json created\n", green, reset)

		return nil
	},
}
