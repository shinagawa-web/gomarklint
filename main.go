package main

import (
	"os"

	"github.com/shinagawa-web/gomarklint/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
