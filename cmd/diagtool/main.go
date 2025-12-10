package main

import (
	"os"

	"github.com/mark/dsl-diagram-tool/cmd/diagtool/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
