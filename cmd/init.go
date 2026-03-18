package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigJSON = `{
  "default": true,
  "rules": {
    "final-blank-line": true,
    "unclosed-code-block": true,
    "empty-alt-text": true,
    "heading-level": { "enabled": true, "severity": "error", "minLevel": 2 },
    "duplicate-heading": true,
    "no-multiple-blank-lines": true,
    "no-setext-headings": true,
    "external-link": { "enabled": false, "severity": "error", "timeoutSeconds": 5, "skipPatterns": [] }
  },
  "include": ["README.md", "docs"],
  "ignore": [],
  "output": "text"
}
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default .gomarklint.json config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println()
		path := ".gomarklint.json"

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		if err := os.WriteFile(path, []byte(defaultConfigJSON), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		green := "\033[32m"
		reset := "\033[0m"

		fmt.Printf("%s✔%s .gomarklint.json created\n", green, reset)

		return nil
	},
}
