package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
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
	watchMode    bool
	pixelDensity int
)

var renderCmd = &cobra.Command{
	Use:   "render <input.d2>",
	Short: "Render a D2 diagram to SVG, PNG, or PDF",
	Long: `Render a D2 diagram file to the specified output format.

Supported output formats:
  - svg (default): Scalable Vector Graphics
  - png: Portable Network Graphics (using headless Chrome)
  - pdf: Portable Document Format (using headless Chrome)

PNG export uses headless Chrome for high-quality conversion with proper font rendering.
The default pixel density is 3x for crisp, high-DPI output. Use --pixel-density to adjust.

The output filename is derived from the input filename if not specified.
For example, 'diagram.d2' will produce 'diagram.svg' by default.

Examples:
  # Render to SVG (default)
  diagtool render diagram.d2

  # Render to PNG (auto-detected from extension)
  diagtool render diagram.d2 -o diagram.png

  # Render to PNG with extra-high resolution
  diagtool render diagram.d2 -o diagram.png --pixel-density 4

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

  # Watch mode: auto-regenerate on file changes
  diagtool render diagram.d2 --watch
  diagtool render diagram.d2 -w -o output.png

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
	renderCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch input file for changes and auto-regenerate")
	renderCmd.Flags().IntVar(&pixelDensity, "pixel-density", 3, "PNG pixel density/DPI multiplier (1=standard, 2=retina, 3-4=high-DPI)")
}

// renderConfig holds the resolved configuration for rendering
type renderConfig struct {
	inputFile string
	outPath   string
	format    string
	opts      render.Options
}

// resolveRenderConfig determines output path and format from flags and input file
func resolveRenderConfig(inputFile string) (*renderConfig, error) {
	// Determine output file path first (to potentially auto-detect format)
	outPath := outputFile

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
		return nil, fmt.Errorf("unsupported output format: %s (use svg, png, or pdf)", format)
	}

	// Derive output path if not specified
	if outPath == "" {
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		outPath = base + "." + format
	}

	// Create render options
	opts := render.Options{
		Format:       render.Format(format),
		ThemeID:      themeID,
		DarkMode:     darkMode,
		Sketch:       sketchMode,
		Padding:      padding,
		Center:       !noCenter,
		Scale:        1.0,
		PixelDensity: pixelDensity,
	}

	return &renderConfig{
		inputFile: inputFile,
		outPath:   outPath,
		format:    format,
		opts:      opts,
	}, nil
}

// doRender performs a single render operation
func doRender(cfg *renderConfig) error {
	// Read input file
	content, err := os.ReadFile(cfg.inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	ctx := context.Background()
	var output []byte

	switch cfg.format {
	case "svg":
		output, err = render.RenderFromSource(ctx, string(content), cfg.opts)
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
		pngRenderer, err := render.NewPNGRendererWithOptions(cfg.opts)
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
		// Parse D2 to IR
		p := parser.NewD2Parser()
		diagram, err := p.Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing failed: %w", err)
		}

		// Create PDF renderer
		pdfRenderer, err := render.NewPDFRendererWithOptions(cfg.opts)
		if err != nil {
			return fmt.Errorf("failed to initialize PDF renderer: %w", err)
		}
		defer pdfRenderer.Close()

		// Render to PDF
		output, err = pdfRenderer.RenderToBytes(ctx, diagram)
		if err != nil {
			return fmt.Errorf("PDF rendering failed: %w", err)
		}
	}

	// Write output file
	if err := os.WriteFile(cfg.outPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func runRender(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Resolve configuration
	cfg, err := resolveRenderConfig(inputFile)
	if err != nil {
		return err
	}

	// Single render mode
	if !watchMode {
		if err := doRender(cfg); err != nil {
			return err
		}
		fmt.Printf("Rendered %s → %s\n", cfg.inputFile, cfg.outPath)
		return nil
	}

	// Watch mode
	return runWatchMode(cfg)
}

// runWatchMode watches the input file and re-renders on changes
func runWatchMode(cfg *renderConfig) error {
	// Get absolute path for reliable watching
	absPath, err := filepath.Abs(cfg.inputFile)
	if err != nil {
		return fmt.Errorf("failed to resolve input path: %w", err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Watch the directory containing the file (more reliable for editor saves)
	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Do initial render
	fmt.Printf("Watching %s for changes (Ctrl+C to stop)...\n", cfg.inputFile)
	if err := doRender(cfg); err != nil {
		fmt.Printf("[%s] Error: %v\n", formatTime(), err)
	} else {
		fmt.Printf("[%s] Rendered %s → %s\n", formatTime(), cfg.inputFile, cfg.outPath)
	}

	// Debounce timer to avoid multiple renders for rapid changes
	var debounceTimer *time.Timer
	const debounceDelay = 100 * time.Millisecond

	baseName := filepath.Base(absPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to changes to our specific file
			if filepath.Base(event.Name) != baseName {
				continue
			}

			// Only react to write/create events (editors may delete+create)
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
				continue
			}

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				if err := doRender(cfg); err != nil {
					fmt.Printf("[%s] Error: %v\n", formatTime(), err)
				} else {
					fmt.Printf("[%s] Rendered %s → %s\n", formatTime(), cfg.inputFile, cfg.outPath)
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("[%s] Watch error: %v\n", formatTime(), err)

		case <-sigChan:
			fmt.Printf("\nStopping watch mode.\n")
			return nil
		}
	}
}

// formatTime returns a formatted timestamp for watch mode output
func formatTime() string {
	return time.Now().Format("15:04:05")
}
