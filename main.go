package main

import (
	"fmt"
	"os"

	"github.com/shinagawa-web/gomarklint/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "[gomarklint error]:", err)
		os.Exit(1)
	}
}
