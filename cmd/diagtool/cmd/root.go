// Package cmd provides the CLI commands for diagtool.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information (set at build time)
var (
	Version   = "1.0.0"
	BuildDate = "2025-12-14"
	GitCommit = "HEAD"
)

var rootCmd = &cobra.Command{
	Use:   "diagtool",
	Short: "DSL Diagram Tool - Render D2 diagrams to various formats",
	Long: `DiagTool is a command-line tool for rendering D2 diagrams to SVG, PNG, and PDF.

It bridges text-based diagram creation with visual output, enabling
version-controlled diagram-as-code workflows.

Examples:
  # Render a D2 file to SVG
  diagtool render diagram.d2 -o diagram.svg

  # Render to PNG with custom options
  diagtool render diagram.d2 -o diagram.png --format png --theme 1

  # Validate a D2 file
  diagtool validate diagram.d2

For more information, visit: https://github.com/mark/dsl-diagram-tool`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
}

// exitWithError prints an error message and exits with code 1.
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
