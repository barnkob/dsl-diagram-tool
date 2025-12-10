package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mark/dsl-diagram-tool/pkg/parser"
)

var validateCmd = &cobra.Command{
	Use:   "validate <input.d2>",
	Short: "Validate a D2 diagram file",
	Long: `Validate a D2 diagram file for syntax errors and structural issues.

This command parses the input file and reports any errors found.
It does not produce any output files.

Examples:
  # Validate a single file
  diagtool validate diagram.d2

  # Validate and show details on success
  diagtool validate diagram.d2 -v`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

var verbose bool

func init() {
	validateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output on success")
}

func runValidate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Read input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse the file
	p := parser.NewD2Parser()
	diagram, err := p.Parse(string(content))
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate the parsed diagram
	validationErrors := diagram.Validate()
	if len(validationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Validation errors in %s:\n", inputFile)
		for _, err := range validationErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		return fmt.Errorf("found %d validation error(s)", len(validationErrors))
	}

	// Success
	if verbose {
		fmt.Printf("✓ %s is valid\n", inputFile)
		fmt.Printf("  Nodes: %d\n", len(diagram.Nodes))
		fmt.Printf("  Edges: %d\n", len(diagram.Edges))
	} else {
		fmt.Printf("✓ %s is valid (%d nodes, %d edges)\n",
			inputFile, len(diagram.Nodes), len(diagram.Edges))
	}

	return nil
}
