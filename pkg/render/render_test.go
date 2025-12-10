package render

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
	"github.com/mark/dsl-diagram-tool/pkg/parser"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Format != FormatSVG {
		t.Errorf("Expected default format SVG, got %s", opts.Format)
	}
	if opts.ThemeID != 0 {
		t.Errorf("Expected default ThemeID 0, got %d", opts.ThemeID)
	}
	if opts.DarkMode {
		t.Error("Expected DarkMode false by default")
	}
	if opts.Sketch {
		t.Error("Expected Sketch false by default")
	}
	if opts.Padding != 100 {
		t.Errorf("Expected default Padding 100, got %d", opts.Padding)
	}
	if !opts.Center {
		t.Error("Expected Center true by default")
	}
	if opts.Scale != 1.0 {
		t.Errorf("Expected default Scale 1.0, got %f", opts.Scale)
	}
}

func TestNewSVGRenderer(t *testing.T) {
	r := NewSVGRenderer()
	if r == nil {
		t.Fatal("NewSVGRenderer returned nil")
	}
	if r.Options.Format != FormatSVG {
		t.Errorf("Expected SVG format, got %s", r.Options.Format)
	}
}

func TestNewSVGRendererWithOptions(t *testing.T) {
	opts := Options{
		ThemeID:  3,
		Padding:  50,
		Sketch:   true,
		DarkMode: true,
	}
	r := NewSVGRendererWithOptions(opts)
	if r == nil {
		t.Fatal("NewSVGRendererWithOptions returned nil")
	}
	if r.Options.ThemeID != 3 {
		t.Errorf("Expected ThemeID 3, got %d", r.Options.ThemeID)
	}
	if r.Options.Padding != 50 {
		t.Errorf("Expected Padding 50, got %d", r.Options.Padding)
	}
	if !r.Options.Sketch {
		t.Error("Expected Sketch true")
	}
	if !r.Options.DarkMode {
		t.Error("Expected DarkMode true")
	}
}

func TestSVGRenderer_RenderToBytes_Simple(t *testing.T) {
	// Create a simple diagram
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{ID: "server", Label: "Web Server", Shape: ir.ShapeRectangle},
			{ID: "database", Label: "Database", Shape: ir.ShapeCylinder},
		},
		Edges: []*ir.Edge{
			{ID: "e1", Source: "server", Target: "database", Direction: ir.DirectionForward, Label: "SQL"},
		},
	}

	r := NewSVGRenderer()
	ctx := context.Background()

	svg, err := r.RenderToBytes(ctx, diagram)
	if err != nil {
		t.Fatalf("RenderToBytes failed: %v", err)
	}

	// Check that output is valid SVG
	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
	if !bytes.Contains(svg, []byte("</svg>")) {
		t.Error("Output doesn't contain closing </svg> tag")
	}

	// Check that content is rendered
	if !bytes.Contains(svg, []byte("Web Server")) {
		t.Error("SVG doesn't contain node label 'Web Server'")
	}
	if !bytes.Contains(svg, []byte("Database")) {
		t.Error("SVG doesn't contain node label 'Database'")
	}
}

func TestSVGRenderer_Render_Writer(t *testing.T) {
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{ID: "a", Label: "Node A", Shape: ir.ShapeRectangle},
		},
		Edges: []*ir.Edge{},
	}

	r := NewSVGRenderer()
	ctx := context.Background()

	var buf bytes.Buffer
	err := r.Render(ctx, diagram, &buf)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	svg := buf.Bytes()
	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
}

func TestRenderFromSource_Simple(t *testing.T) {
	source := `
server: Web Server
database: Database
server -> database: SQL
`
	ctx := context.Background()
	opts := DefaultOptions()

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
	if !bytes.Contains(svg, []byte("Web Server")) {
		t.Error("SVG doesn't contain 'Web Server'")
	}
}

func TestRenderFromSource_WithContainers(t *testing.T) {
	source := `
aws: AWS Cloud {
  vpc: VPC {
    server: Web Server
  }
}
client: Client
client -> aws.vpc.server: API
`
	ctx := context.Background()
	opts := DefaultOptions()

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
	if !bytes.Contains(svg, []byte("AWS Cloud")) {
		t.Error("SVG doesn't contain 'AWS Cloud'")
	}
}

func TestRenderFromSource_WithStyles(t *testing.T) {
	source := `
server: Server {
  style: {
    fill: "#4CAF50"
    stroke: "#2E7D32"
    font-color: white
  }
}
`
	ctx := context.Background()
	opts := DefaultOptions()

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
}

func TestRenderFromSource_SketchMode(t *testing.T) {
	source := `a -> b -> c`
	ctx := context.Background()
	opts := DefaultOptions()
	opts.Sketch = true

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource with sketch mode failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
}

func TestRenderFromSource_DarkMode(t *testing.T) {
	source := `a -> b`
	ctx := context.Background()
	opts := DefaultOptions()
	opts.DarkMode = true

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource with dark mode failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
}

func TestRenderFromSource_CustomPadding(t *testing.T) {
	source := `a -> b`
	ctx := context.Background()
	opts := DefaultOptions()
	opts.Padding = 200

	svg, err := RenderFromSource(ctx, source, opts)
	if err != nil {
		t.Fatalf("RenderFromSource with custom padding failed: %v", err)
	}

	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
}

func TestIrToD2Source_Simple(t *testing.T) {
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{ID: "server", Label: "Web Server", Shape: ir.ShapeRectangle},
			{ID: "database", Label: "Database", Shape: ir.ShapeCylinder},
		},
		Edges: []*ir.Edge{
			{ID: "e1", Source: "server", Target: "database", Direction: ir.DirectionForward, Label: "SQL"},
		},
	}

	source := irToD2Source(diagram)

	if source == "" {
		t.Error("Generated D2 source is empty")
	}
	if !strings.Contains(source, "direction: down") {
		t.Error("Missing direction directive")
	}
	if !strings.Contains(source, "server") {
		t.Error("Missing server node")
	}
	if !strings.Contains(source, "database") {
		t.Error("Missing database node")
	}
	if !strings.Contains(source, "->") {
		t.Error("Missing edge arrow")
	}
}

func TestIrToD2Source_WithContainers(t *testing.T) {
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{ID: "aws", Label: "AWS", Shape: ir.ShapeContainer},
			{ID: "aws.server", Label: "Server", Shape: ir.ShapeRectangle, Container: "aws"},
		},
		Edges: []*ir.Edge{},
	}

	source := irToD2Source(diagram)

	if !strings.Contains(source, "aws") {
		t.Error("Missing aws container")
	}
	if !strings.Contains(source, "server") {
		t.Error("Missing server node")
	}
}

func TestIrToD2Source_WithStyles(t *testing.T) {
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{
				ID:    "styled",
				Label: "Styled Node",
				Shape: ir.ShapeRectangle,
				Style: ir.Style{
					Fill:      "#4CAF50",
					Stroke:    "#2E7D32",
					FontColor: "white",
					Bold:      true,
				},
			},
		},
		Edges: []*ir.Edge{},
	}

	source := irToD2Source(diagram)

	if !strings.Contains(source, "style:") {
		t.Error("Missing style block")
	}
	if !strings.Contains(source, "#4CAF50") {
		t.Error("Missing fill color")
	}
	if !strings.Contains(source, "#2E7D32") {
		t.Error("Missing stroke color")
	}
	if !strings.Contains(source, "bold: true") {
		t.Error("Missing bold style")
	}
}

func TestIrToD2SourceWithDirection(t *testing.T) {
	diagram := &ir.Diagram{
		ID:    "test",
		Nodes: []*ir.Node{{ID: "a", Shape: ir.ShapeRectangle}},
	}

	tests := []struct {
		direction string
		expected  string
	}{
		{"down", "direction: down"},
		{"up", "direction: up"},
		{"right", "direction: right"},
		{"left", "direction: left"},
	}

	for _, tt := range tests {
		t.Run(tt.direction, func(t *testing.T) {
			source := irToD2SourceWithDirection(diagram, tt.direction)
			if !strings.Contains(source, tt.expected) {
				t.Errorf("Expected %q in source, got: %s", tt.expected, source)
			}
		})
	}
}

func TestShapeToD2(t *testing.T) {
	tests := []struct {
		shape    ir.ShapeType
		expected string
	}{
		{ir.ShapeRectangle, "rectangle"},
		{ir.ShapePerson, "person"},
		{ir.ShapeCloud, "cloud"},
		{ir.ShapeCylinder, "cylinder"},
		{ir.ShapeCircle, "circle"},
		{ir.ShapeDiamond, "diamond"},
		{ir.ShapeHexagon, "hexagon"},
		{ir.ShapeSquare, "square"},
		{ir.ShapeParallelogram, "parallelogram"},
		{ir.ShapeSQLTable, "sql_table"},
		{ir.ShapeClass, "class"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shape), func(t *testing.T) {
			result := shapeToD2(tt.shape)
			if result != tt.expected {
				t.Errorf("shapeToD2(%s) = %s, expected %s", tt.shape, result, tt.expected)
			}
		})
	}
}

func TestHasNonDefaultStyle(t *testing.T) {
	tests := []struct {
		name     string
		style    ir.Style
		expected bool
	}{
		{"empty style", ir.Style{}, false},
		{"with fill", ir.Style{Fill: "#fff"}, true},
		{"with stroke", ir.Style{Stroke: "#000"}, true},
		{"with bold", ir.Style{Bold: true}, true},
		{"with shadow", ir.Style{Shadow: true}, true},
		{"with animated", ir.Style{Animated: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasNonDefaultStyle(tt.style)
			if result != tt.expected {
				t.Errorf("hasNonDefaultStyle() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestWriteEdge(t *testing.T) {
	tests := []struct {
		name     string
		edge     *ir.Edge
		expected string
	}{
		{
			"forward arrow",
			&ir.Edge{Source: "a", Target: "b", Direction: ir.DirectionForward},
			"a -> b\n",
		},
		{
			"backward arrow",
			&ir.Edge{Source: "a", Target: "b", Direction: ir.DirectionBackward},
			"a <- b\n",
		},
		{
			"bidirectional",
			&ir.Edge{Source: "a", Target: "b", Direction: ir.DirectionBoth},
			"a <-> b\n",
		},
		{
			"no arrow",
			&ir.Edge{Source: "a", Target: "b", Direction: ir.DirectionNone},
			"a -- b\n",
		},
		{
			"with label",
			&ir.Edge{Source: "a", Target: "b", Direction: ir.DirectionForward, Label: "connects to"},
			"a -> b: connects to\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writeEdge(tt.edge)
			if result != tt.expected {
				t.Errorf("writeEdge() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Integration test: Parse -> Render roundtrip
func TestParseAndRender_Roundtrip(t *testing.T) {
	source := `
server: Web Server
database: Database { shape: cylinder }
cache: Cache { shape: circle }

server -> database: queries
server -> cache: reads
`
	// Parse
	p := parser.NewD2Parser()
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Render
	r := NewSVGRenderer()
	ctx := context.Background()
	svg, err := r.RenderToBytes(ctx, diagram)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify SVG output
	if !bytes.Contains(svg, []byte("<svg")) {
		t.Error("Output doesn't contain <svg tag")
	}
	if !bytes.Contains(svg, []byte("Web Server")) {
		t.Error("SVG doesn't contain 'Web Server'")
	}
	if !bytes.Contains(svg, []byte("Database")) {
		t.Error("SVG doesn't contain 'Database'")
	}
}

// Test rendering example files
func TestRender_ExampleFiles(t *testing.T) {
	examplesDir := "../../examples"

	files, err := filepath.Glob(filepath.Join(examplesDir, "*.d2"))
	if err != nil {
		t.Fatalf("Failed to glob examples: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No example files found")
	}

	ctx := context.Background()
	opts := DefaultOptions()

	for _, file := range files {
		// Skip macOS metadata files
		if strings.HasPrefix(filepath.Base(file), "._") {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			svg, err := RenderFromSource(ctx, string(content), opts)
			if err != nil {
				t.Fatalf("Failed to render %s: %v", file, err)
			}

			if !bytes.Contains(svg, []byte("<svg")) {
				t.Error("Output doesn't contain <svg tag")
			}

			t.Logf("Rendered %s: %d bytes SVG", filepath.Base(file), len(svg))
		})
	}
}

// Benchmark tests
func BenchmarkRenderFromSource_Simple(b *testing.B) {
	source := `
a -> b -> c -> d
d -> e -> f
`
	ctx := context.Background()
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderFromSource(ctx, source, opts)
	}
}

func BenchmarkRenderFromSource_Complex(b *testing.B) {
	source := `
aws: AWS {
  vpc: VPC {
    web1: Web 1
    web2: Web 2
    app1: App 1
    app2: App 2
    db: Database { shape: cylinder }
  }
}
web1 -> app1
web2 -> app2
app1 -> db
app2 -> db
`
	ctx := context.Background()
	opts := DefaultOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RenderFromSource(ctx, source, opts)
	}
}

func BenchmarkSVGRenderer_RenderToBytes(b *testing.B) {
	diagram := &ir.Diagram{
		ID: "test",
		Nodes: []*ir.Node{
			{ID: "a", Label: "Node A", Shape: ir.ShapeRectangle},
			{ID: "b", Label: "Node B", Shape: ir.ShapeRectangle},
			{ID: "c", Label: "Node C", Shape: ir.ShapeRectangle},
		},
		Edges: []*ir.Edge{
			{ID: "e1", Source: "a", Target: "b", Direction: ir.DirectionForward},
			{ID: "e2", Source: "b", Target: "c", Direction: ir.DirectionForward},
		},
	}

	r := NewSVGRenderer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RenderToBytes(ctx, diagram)
	}
}
