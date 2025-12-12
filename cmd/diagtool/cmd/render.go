package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mark/dsl-diagram-tool/pkg/parser"
	"github.com/mark/dsl-diagram-tool/pkg/render"
)

var (
	outputFile   string
	outputFormat string
	themeID      int64
	darkMode     bool
	sketchMode   bool
	padding      int64
	noCenter     bool
)

var renderCmd = &cobra.Command{
	Use:   "render <input.d2>",
	Short: "Render a D2 diagram to SVG, PNG, or PDF",
	Long: `Render a D2 diagram file to the specified output format.

Supported output formats:
  - svg (default): Scalable Vector Graphics
  - png: Portable Network Graphics (using resvg)
  - pdf: Portable Document Format (not yet implemented)

PNG export uses resvg for high-quality conversion with no external dependencies.

The output filename is derived from the input filename if not specified.
For example, 'diagram.d2' will produce 'diagram.svg' by default.

Examples:
  # Render to SVG (default)
  diagtool render diagram.d2

  # Render to PNG (auto-detected from extension)
  diagtool render diagram.d2 -o diagram.png

  # Render to PNG (explicit format)
  diagtool render diagram.d2 -f png

  # Specify output file
  diagtool render diagram.d2 -o output.svg

  # Use dark theme
  diagtool render diagram.d2 --dark

  # Use sketch/hand-drawn style
  diagtool render diagram.d2 --sketch

  # Use a specific theme (0-8)
  diagtool render diagram.d2 --theme 3

Note: Format is auto-detected from output file extension (.png, .svg, .pdf).
Use -f to explicitly override the format.`,
	Args: cobra.ExactArgs(1),
	RunE: runRender,
}

func init() {
	renderCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: input name with format extension)")
	renderCmd.Flags().StringVarP(&outputFormat, "format", "f", "svg", "Output format: svg, png, pdf")
	renderCmd.Flags().Int64VarP(&themeID, "theme", "t", 0, "Theme ID (0-8, default: 0)")
	renderCmd.Flags().BoolVarP(&darkMode, "dark", "d", false, "Use dark mode theme")
	renderCmd.Flags().BoolVarP(&sketchMode, "sketch", "s", false, "Use sketch/hand-drawn style")
	renderCmd.Flags().Int64VarP(&padding, "padding", "p", 100, "Padding around diagram in pixels")
	renderCmd.Flags().BoolVar(&noCenter, "no-center", false, "Don't center the diagram")
}

func runRender(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Read input file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Determine output file path first (to potentially auto-detect format)
	outPath := outputFile
	if outPath == "" {
		// Will derive after we know the format
		outPath = ""
	}

	// Determine output format
	// Auto-detect from output file extension if -f not specified
	format := strings.ToLower(outputFormat)
	if format == "svg" && outPath != "" {
		// Check if user specified a different extension (auto-detect)
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(outPath), "."))
		if ext == "png" || ext == "pdf" {
			format = ext
		}
	}

	// Validate format
	switch format {
	case "svg", "png", "pdf":
		// Valid format
	default:
		return fmt.Errorf("unsupported output format: %s (use svg, png, or pdf)", format)
	}

	// Derive output path if not specified
	if outPath == "" {
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outPath = base + "." + format
	}

	// Create render options
	opts := render.Options{
		Format:   render.Format(format),
		ThemeID:  themeID,
		DarkMode: darkMode,
		Sketch:   sketchMode,
		Padding:  padding,
		Center:   !noCenter,
		Scale:    1.0,
	}

	// Render the diagram
	ctx := context.Background()
	var output []byte

	switch format {
	case "svg":
		output, err = render.RenderFromSource(ctx, string(content), opts)
		if err != nil {
			return fmt.Errorf("rendering failed: %w", err)
		}
	case "png":
		// Parse D2 to IR
		p := parser.NewD2Parser()
		diagram, err := p.Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing failed: %w", err)
		}

		// Create PNG renderer
		pngRenderer, err := render.NewPNGRendererWithOptions(opts)
		if err != nil {
			return fmt.Errorf("failed to initialize PNG renderer: %w", err)
		}
		defer pngRenderer.Close()

		// Render to PNG
		output, err = pngRenderer.RenderToBytes(ctx, diagram)
		if err != nil {
			return fmt.Errorf("PNG rendering failed: %w", err)
		}
	case "pdf":
		// PDF export not yet implemented
		return fmt.Errorf("PDF export is not yet implemented (use svg or png for now)")
	}

	// Write output file
	if err := os.WriteFile(outPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Rendered %s → %s (%d bytes)\n", inputFile, outPath, len(output))
	return nil
}
