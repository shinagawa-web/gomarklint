package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/shinagawa-web/gomarklint/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Don't print error message for lint violations (already displayed)
		if !errors.Is(err, cmd.ErrLintViolations) {
			fmt.Fprintln(os.Stderr, "[gomarklint error]:", err)
		}
		os.Exit(1)
	}
}
