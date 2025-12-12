// Package render provides diagram rendering to various formats.
// This wraps D2's rendering capabilities (SVG, PNG, PDF) and provides
// a unified interface for diagram output.
package render

import (
	"context"
	"fmt"
	"io"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/png"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
)

// Format represents the output format for rendering.
type Format string

// Supported output formats.
const (
	FormatSVG Format = "svg"
	FormatPNG Format = "png"
	FormatPDF Format = "pdf"
)

// Options configures the rendering behavior.
type Options struct {
	// Output format (default: SVG)
	Format Format

	// Theme ID (default: 0, which is the default D2 theme)
	// Other themes: 1-8 are built-in D2 themes
	ThemeID int64

	// Dark mode (default: false)
	DarkMode bool

	// Sketch mode - hand-drawn appearance (default: false)
	Sketch bool

	// Padding around the diagram in pixels (default: 100)
	Padding int64

	// Center the diagram in the viewport (default: true)
	Center bool

	// Scale factor for rendering (default: 1.0)
	// Values > 1 produce larger output, < 1 produce smaller
	Scale float64

	// For PNG: pixel density (default: 2 for retina)
	PixelDensity int
}

// DefaultOptions returns sensible default rendering options.
func DefaultOptions() Options {
	return Options{
		Format:       FormatSVG,
		ThemeID:      0,
		DarkMode:     false,
		Sketch:       false,
		Padding:      100,
		Center:       true,
		Scale:        1.0,
		PixelDensity: 2,
	}
}

// Renderer is the interface for diagram renderers.
type Renderer interface {
	// Render renders the diagram to the provided writer.
	Render(ctx context.Context, diagram *ir.Diagram, w io.Writer) error

	// RenderToBytes renders the diagram and returns the output as bytes.
	RenderToBytes(ctx context.Context, diagram *ir.Diagram) ([]byte, error)
}

// SVGRenderer renders diagrams to SVG format using D2's rendering engine.
type SVGRenderer struct {
	Options Options
}

// NewSVGRenderer creates a new SVG renderer with default options.
func NewSVGRenderer() *SVGRenderer {
	opts := DefaultOptions()
	opts.Format = FormatSVG
	return &SVGRenderer{Options: opts}
}

// NewSVGRendererWithOptions creates a new SVG renderer with custom options.
func NewSVGRendererWithOptions(opts Options) *SVGRenderer {
	opts.Format = FormatSVG
	return &SVGRenderer{Options: opts}
}

// PNGRenderer renders diagrams to PNG format using D2's rendering pipeline.
// This uses playwright under the hood to convert SVG to PNG.
type PNGRenderer struct {
	Options    Options
	playwright png.Playwright
}

// NewPNGRenderer creates a new PNG renderer with default options.
// Initializes playwright for SVG to PNG conversion.
func NewPNGRenderer() (*PNGRenderer, error) {
	opts := DefaultOptions()
	opts.Format = FormatPNG

	pw, err := png.InitPlaywright()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playwright: %w", err)
	}

	return &PNGRenderer{
		Options:    opts,
		playwright: pw,
	}, nil
}

// NewPNGRendererWithOptions creates a new PNG renderer with custom options.
func NewPNGRendererWithOptions(opts Options) (*PNGRenderer, error) {
	opts.Format = FormatPNG

	pw, err := png.InitPlaywright()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playwright: %w", err)
	}

	return &PNGRenderer{
		Options:    opts,
		playwright: pw,
	}, nil
}

// Close releases playwright resources. Should be called when done rendering.
func (r *PNGRenderer) Close() error {
	if r.playwright.Browser != nil {
		return r.playwright.Browser.Close()
	}
	return nil
}

// Render renders the diagram to PNG format.
func (r *PNGRenderer) Render(ctx context.Context, diagram *ir.Diagram, w io.Writer) error {
	pngBytes, err := r.RenderToBytes(ctx, diagram)
	if err != nil {
		return err
	}
	_, err = w.Write(pngBytes)
	return err
}

// RenderToBytes renders the diagram and returns PNG as bytes.
func (r *PNGRenderer) RenderToBytes(ctx context.Context, diagram *ir.Diagram) ([]byte, error) {
	// First render to SVG
	svgRenderer := NewSVGRendererWithOptions(r.Options)
	svgBytes, err := svgRenderer.RenderToBytes(ctx, diagram)
	if err != nil {
		return nil, fmt.Errorf("failed to render SVG for PNG conversion: %w", err)
	}

	// Convert SVG to PNG using playwright
	page, err := r.playwright.Browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser page: %w", err)
	}
	defer page.Close()

	pngBytes, err := png.ConvertSVG(page, svgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SVG to PNG: %w", err)
	}

	// Add EXIF metadata
	pngBytes, err = png.AddExif(pngBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to add EXIF metadata: %w", err)
	}

	return pngBytes, nil
}

// Render renders the diagram to SVG format.
func (r *SVGRenderer) Render(ctx context.Context, diagram *ir.Diagram, w io.Writer) error {
	svgBytes, err := r.RenderToBytes(ctx, diagram)
	if err != nil {
		return err
	}
	_, err = w.Write(svgBytes)
	return err
}

// RenderToBytes renders the diagram and returns SVG as bytes.
func (r *SVGRenderer) RenderToBytes(ctx context.Context, diagram *ir.Diagram) ([]byte, error) {
	// Convert IR to D2 source
	d2Source := irToD2Source(diagram)

	// Create text ruler for measurement
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("failed to create text ruler: %w", err)
	}

	// Create layout resolver
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return func(ctx context.Context, g *d2graph.Graph) error {
			return d2dagrelayout.Layout(ctx, g, nil)
		}, nil
	}

	// Compile options
	compileOpts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}

	// Render options
	renderOpts := &d2svg.RenderOpts{
		ThemeID: &r.Options.ThemeID,
		Pad:     &r.Options.Padding,
		Sketch:  &r.Options.Sketch,
		Center:  &r.Options.Center,
	}

	if r.Options.DarkMode {
		darkThemeID := r.Options.ThemeID + 100 // D2 dark themes are offset by 100
		renderOpts.ThemeID = &darkThemeID
	}

	// Compile the diagram
	targetDiagram, _, err := d2lib.Compile(ctx, d2Source, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Render to SVG
	svg, err := d2svg.Render(targetDiagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("SVG rendering failed: %w", err)
	}

	return svg, nil
}

// RenderFromSource renders D2 source directly to SVG.
// This is more efficient when you have the original D2 source.
func RenderFromSource(ctx context.Context, source string, opts Options) ([]byte, error) {
	// Create text ruler for measurement
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("failed to create text ruler: %w", err)
	}

	// Create layout resolver
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return func(ctx context.Context, g *d2graph.Graph) error {
			return d2dagrelayout.Layout(ctx, g, nil)
		}, nil
	}

	// Compile options
	compileOpts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}

	// Render options
	renderOpts := &d2svg.RenderOpts{
		ThemeID: &opts.ThemeID,
		Pad:     &opts.Padding,
		Sketch:  &opts.Sketch,
		Center:  &opts.Center,
	}

	if opts.DarkMode {
		darkThemeID := opts.ThemeID + 100
		renderOpts.ThemeID = &darkThemeID
	}

	// Compile
	targetDiagram, _, err := d2lib.Compile(ctx, source, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Render
	svg, err := d2svg.Render(targetDiagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("SVG rendering failed: %w", err)
	}

	return svg, nil
}

// irToD2Source converts an IR diagram to D2 source code for rendering.
func irToD2Source(diagram *ir.Diagram) string {
	return irToD2SourceWithDirection(diagram, "down")
}

// irToD2SourceWithDirection converts IR to D2 with a specified direction.
func irToD2SourceWithDirection(diagram *ir.Diagram, direction string) string {
	var result string

	// Add direction directive
	result += fmt.Sprintf("direction: %s\n\n", direction)

	// Track containers
	containers := make(map[string]bool)
	for _, node := range diagram.Nodes {
		if node.Shape == ir.ShapeContainer {
			containers[node.ID] = true
		}
	}

	// Write root-level nodes
	for _, node := range diagram.Nodes {
		if node.Container == "" && node.GetParentID() == "" {
			result += writeNode(node, diagram, containers, 0)
		}
	}

	result += "\n"

	// Write edges
	for _, edge := range diagram.Edges {
		result += writeEdge(edge)
	}

	return result
}

// writeNode writes a node and its children to D2 format.
func writeNode(node *ir.Node, diagram *ir.Diagram, containers map[string]bool, indent int) string {
	var result string
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	// Get local ID
	localID := node.ID
	for i := len(node.ID) - 1; i >= 0; i-- {
		if node.ID[i] == '.' {
			localID = node.ID[i+1:]
			break
		}
	}

	// Node declaration
	if node.Label != "" && node.Label != localID {
		result += fmt.Sprintf("%s%s: %s", prefix, localID, node.Label)
	} else {
		result += fmt.Sprintf("%s%s", prefix, localID)
	}

	// Check if container or has styling
	isContainer := containers[node.ID]
	hasShape := node.Shape != ir.ShapeRectangle && node.Shape != ir.ShapeContainer
	hasStyle := hasNonDefaultStyle(node.Style)

	if isContainer || hasShape || hasStyle {
		result += " {\n"

		// Shape
		if hasShape {
			result += fmt.Sprintf("%s  shape: %s\n", prefix, shapeToD2(node.Shape))
		}

		// Styling
		if hasStyle {
			result += writeStyle(node.Style, prefix+"  ")
		}

		// Children
		if isContainer {
			children := diagram.GetNodesByContainer(node.ID)
			for _, child := range children {
				result += writeNode(child, diagram, containers, indent+1)
			}
		}

		result += fmt.Sprintf("%s}\n", prefix)
	} else {
		result += "\n"
	}

	return result
}

// writeEdge writes an edge in D2 format.
func writeEdge(edge *ir.Edge) string {
	arrow := "->"
	switch edge.Direction {
	case ir.DirectionBackward:
		arrow = "<-"
	case ir.DirectionBoth:
		arrow = "<->"
	case ir.DirectionNone:
		arrow = "--"
	}

	if edge.Label != "" {
		return fmt.Sprintf("%s %s %s: %s\n", edge.Source, arrow, edge.Target, edge.Label)
	}
	return fmt.Sprintf("%s %s %s\n", edge.Source, arrow, edge.Target)
}

// shapeToD2 converts IR shape type to D2 shape string.
func shapeToD2(shape ir.ShapeType) string {
	switch shape {
	case ir.ShapePerson:
		return "person"
	case ir.ShapeCloud:
		return "cloud"
	case ir.ShapeCylinder:
		return "cylinder"
	case ir.ShapeCircle:
		return "circle"
	case ir.ShapeOval:
		return "oval"
	case ir.ShapeDiamond:
		return "diamond"
	case ir.ShapeHexagon:
		return "hexagon"
	case ir.ShapeSquare:
		return "square"
	case ir.ShapeParallelogram:
		return "parallelogram"
	case ir.ShapeSQLTable:
		return "sql_table"
	case ir.ShapeClass:
		return "class"
	default:
		return "rectangle"
	}
}

// hasNonDefaultStyle returns true if the style has any non-default values.
func hasNonDefaultStyle(s ir.Style) bool {
	return s.Fill != "" || s.Stroke != "" || s.StrokeWidth != 0 ||
		s.StrokeDash != 0 || s.BorderRadius != 0 || s.Opacity != 0 ||
		s.Shadow || s.ThreeD || s.Multiple || s.DoubleBorder ||
		s.Font != "" || s.FontSize != 0 || s.FontColor != "" ||
		s.Bold || s.Italic || s.Underline || s.TextTransform != "" ||
		s.Animated
}

// writeStyle writes style block to D2 format.
func writeStyle(s ir.Style, prefix string) string {
	var result string
	result += prefix + "style: {\n"

	if s.Fill != "" {
		result += fmt.Sprintf("%s  fill: \"%s\"\n", prefix, s.Fill)
	}
	if s.Stroke != "" {
		result += fmt.Sprintf("%s  stroke: \"%s\"\n", prefix, s.Stroke)
	}
	if s.StrokeWidth != 0 {
		result += fmt.Sprintf("%s  stroke-width: %d\n", prefix, s.StrokeWidth)
	}
	if s.StrokeDash != 0 {
		result += fmt.Sprintf("%s  stroke-dash: %d\n", prefix, s.StrokeDash)
	}
	if s.BorderRadius != 0 {
		result += fmt.Sprintf("%s  border-radius: %d\n", prefix, s.BorderRadius)
	}
	if s.Opacity != 0 {
		result += fmt.Sprintf("%s  opacity: %.2f\n", prefix, s.Opacity)
	}
	if s.Shadow {
		result += fmt.Sprintf("%s  shadow: true\n", prefix)
	}
	if s.ThreeD {
		result += fmt.Sprintf("%s  3d: true\n", prefix)
	}
	if s.Multiple {
		result += fmt.Sprintf("%s  multiple: true\n", prefix)
	}
	if s.DoubleBorder {
		result += fmt.Sprintf("%s  double-border: true\n", prefix)
	}
	if s.Font != "" {
		result += fmt.Sprintf("%s  font: \"%s\"\n", prefix, s.Font)
	}
	if s.FontSize != 0 {
		result += fmt.Sprintf("%s  font-size: %d\n", prefix, s.FontSize)
	}
	if s.FontColor != "" {
		result += fmt.Sprintf("%s  font-color: \"%s\"\n", prefix, s.FontColor)
	}
	if s.Bold {
		result += fmt.Sprintf("%s  bold: true\n", prefix)
	}
	if s.Italic {
		result += fmt.Sprintf("%s  italic: true\n", prefix)
	}
	if s.Underline {
		result += fmt.Sprintf("%s  underline: true\n", prefix)
	}
	if s.Animated {
		result += fmt.Sprintf("%s  animated: true\n", prefix)
	}

	result += prefix + "}\n"
	return result
}
