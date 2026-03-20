package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/shinagawa-web/gomarklint/v2/cmd"
	"github.com/shinagawa-web/gomarklint/v2/internal/app"
)

var osExit = os.Exit

func main() {
	if err := cmd.Execute(); err != nil {
		// Don't print error message for lint violations (already displayed)
		if !errors.Is(err, app.ErrLintViolations) {
			fmt.Fprintln(os.Stderr, "[gomarklint error]:", err)
		}
		osExit(1)
	}
}
