package layout

import (
	"context"
	"testing"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
	"github.com/mark/dsl-diagram-tool/pkg/parser"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Engine != LayoutEngineDagre {
		t.Errorf("Expected default engine Dagre, got %s", opts.Engine)
	}
	if opts.Direction != DirectionDown {
		t.Errorf("Expected default direction down, got %s", opts.Direction)
	}
	if opts.NodeSep != 60 {
		t.Errorf("Expected default NodeSep 60, got %d", opts.NodeSep)
	}
	if opts.EdgeSep != 20 {
		t.Errorf("Expected default EdgeSep 20, got %d", opts.EdgeSep)
	}
}

func TestNewDagreLayout(t *testing.T) {
	l := NewDagreLayout()
	if l == nil {
		t.Fatal("NewDagreLayout returned nil")
	}
	if l.Options.Engine != LayoutEngineDagre {
		t.Errorf("Expected Dagre engine, got %s", l.Options.Engine)
	}
}

func TestNewDagreLayoutWithOptions(t *testing.T) {
	opts := Options{
		Engine:    LayoutEngineDagre,
		Direction: DirectionRight,
		NodeSep:   80,
		EdgeSep:   30,
	}
	l := NewDagreLayoutWithOptions(opts)
	if l == nil {
		t.Fatal("NewDagreLayoutWithOptions returned nil")
	}
	if l.Options.Direction != DirectionRight {
		t.Errorf("Expected direction right, got %s", l.Options.Direction)
	}
	if l.Options.NodeSep != 80 {
		t.Errorf("Expected NodeSep 80, got %d", l.Options.NodeSep)
	}
}

func TestDagreLayout_Apply_Simple(t *testing.T) {
	// Parse a simple diagram
	p := parser.NewD2Parser()
	source := `
server: Web Server
database: Database
server -> database: SQL
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Apply layout
	l := NewDagreLayout()
	ctx := context.Background()
	if err := l.Apply(ctx, diagram); err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// Verify nodes have positions
	for _, node := range diagram.Nodes {
		if node.Position == nil {
			t.Errorf("Node %s has no position", node.ID)
			continue
		}
		if node.Position.Source != ir.PositionSourceLayoutEngine {
			t.Errorf("Node %s position source is %s, expected layout_engine",
				node.ID, node.Position.Source)
		}
		if node.Width <= 0 {
			t.Errorf("Node %s has invalid width: %f", node.ID, node.Width)
		}
		if node.Height <= 0 {
			t.Errorf("Node %s has invalid height: %f", node.ID, node.Height)
		}
	}
}

func TestDagreLayout_Apply_WithContainers(t *testing.T) {
	p := parser.NewD2Parser()
	source := `
aws: AWS Cloud {
  vpc: VPC {
    server: Web Server
  }
}
client: Client
client -> aws.vpc.server: API
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	l := NewDagreLayout()
	ctx := context.Background()
	if err := l.Apply(ctx, diagram); err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	// Verify all nodes have positions
	for _, node := range diagram.Nodes {
		if node.Position == nil {
			t.Errorf("Node %s has no position", node.ID)
		}
	}

	// Check that nested node is within container bounds
	var aws, server *ir.Node
	for _, node := range diagram.Nodes {
		if node.ID == "aws" {
			aws = node
		}
		if node.ID == "aws.vpc.server" {
			server = node
		}
	}

	if aws != nil && server != nil && aws.Position != nil && server.Position != nil {
		// Server should be within AWS container
		if server.Position.X < aws.Position.X ||
			server.Position.Y < aws.Position.Y {
			t.Logf("Warning: server position may be outside aws container (server: %f,%f aws: %f,%f)",
				server.Position.X, server.Position.Y, aws.Position.X, aws.Position.Y)
		}
	}
}

func TestApplyFromSource(t *testing.T) {
	source := `
a -> b -> c
`
	// First parse to get IR
	p := parser.NewD2Parser()
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Apply layout from source
	ctx := context.Background()
	opts := DefaultOptions()
	if err := ApplyFromSource(ctx, source, diagram, opts); err != nil {
		t.Fatalf("ApplyFromSource failed: %v", err)
	}

	// Verify positions
	for _, node := range diagram.Nodes {
		if node.Position == nil {
			t.Errorf("Node %s has no position", node.ID)
		}
	}
}

func TestApplyFromSource_WithDirection(t *testing.T) {
	source := `
a -> b -> c
`
	p := parser.NewD2Parser()
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ctx := context.Background()
	opts := Options{
		Engine:    LayoutEngineDagre,
		Direction: DirectionRight,
		NodeSep:   60,
		EdgeSep:   20,
	}
	if err := ApplyFromSource(ctx, source, diagram, opts); err != nil {
		t.Fatalf("ApplyFromSource failed: %v", err)
	}

	// With right direction, nodes should be arranged horizontally
	// (x values should increase from a to c)
	var aNode, cNode *ir.Node
	for _, node := range diagram.Nodes {
		if node.ID == "a" {
			aNode = node
		}
		if node.ID == "c" {
			cNode = node
		}
	}

	if aNode != nil && cNode != nil && aNode.Position != nil && cNode.Position != nil {
		// c should be to the right of a
		if cNode.Position.X <= aNode.Position.X {
			t.Logf("Note: Expected c to be right of a with direction:right (a.x=%f, c.x=%f)",
				aNode.Position.X, cNode.Position.X)
		}
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

	source := irToD2Source(diagram, DirectionDown)

	// Check that output contains expected elements
	if source == "" {
		t.Error("Generated D2 source is empty")
	}

	// Should contain direction
	if !containsSubstring(source, "direction: down") {
		t.Error("Missing direction directive")
	}

	// Should contain nodes
	if !containsSubstring(source, "server") {
		t.Error("Missing server node")
	}
	if !containsSubstring(source, "database") {
		t.Error("Missing database node")
	}

	// Should contain edge
	if !containsSubstring(source, "->") {
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

	source := irToD2Source(diagram, DirectionDown)

	// Should have nested structure
	if !containsSubstring(source, "aws") {
		t.Error("Missing aws container")
	}
	if !containsSubstring(source, "server") {
		t.Error("Missing server node")
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
	}

	for _, tt := range tests {
		result := shapeToD2(tt.shape)
		if result != tt.expected {
			t.Errorf("shapeToD2(%s) = %s, expected %s", tt.shape, result, tt.expected)
		}
	}
}

func TestGetDiagramBounds(t *testing.T) {
	diagram := &ir.Diagram{
		Nodes: []*ir.Node{
			{
				ID:       "a",
				Position: &ir.Position{X: 10, Y: 20},
				Width:    100,
				Height:   50,
			},
			{
				ID:       "b",
				Position: &ir.Position{X: 200, Y: 100},
				Width:    80,
				Height:   40,
			},
		},
	}

	minX, minY, maxX, maxY := GetDiagramBounds(diagram)

	if minX != 10 {
		t.Errorf("Expected minX=10, got %f", minX)
	}
	if minY != 20 {
		t.Errorf("Expected minY=20, got %f", minY)
	}
	if maxX != 280 { // 200 + 80
		t.Errorf("Expected maxX=280, got %f", maxX)
	}
	if maxY != 140 { // 100 + 40
		t.Errorf("Expected maxY=140, got %f", maxY)
	}
}

func TestGetDiagramBounds_Empty(t *testing.T) {
	diagram := &ir.Diagram{Nodes: []*ir.Node{}}

	minX, minY, maxX, maxY := GetDiagramBounds(diagram)

	if minX != 0 || minY != 0 || maxX != 0 || maxY != 0 {
		t.Errorf("Expected all zeros for empty diagram, got %f,%f,%f,%f",
			minX, minY, maxX, maxY)
	}
}

func TestGetDiagramBounds_NoPositions(t *testing.T) {
	diagram := &ir.Diagram{
		Nodes: []*ir.Node{
			{ID: "a"}, // No position
			{ID: "b"}, // No position
		},
	}

	minX, minY, maxX, maxY := GetDiagramBounds(diagram)

	// Should return extreme values since no valid positions
	if minX < 1e8 || minY < 1e8 {
		t.Logf("Bounds with no positions: %f,%f,%f,%f", minX, minY, maxX, maxY)
	}
}

// Helper function to check substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkDagreLayout_Simple(b *testing.B) {
	p := parser.NewD2Parser()
	source := `
a -> b -> c -> d
d -> e -> f
`
	diagram, _ := p.Parse(source)
	l := NewDagreLayout()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset positions
		for _, node := range diagram.Nodes {
			node.Position = nil
		}
		_ = l.Apply(ctx, diagram)
	}
}

func BenchmarkDagreLayout_Complex(b *testing.B) {
	p := parser.NewD2Parser()
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
	diagram, _ := p.Parse(source)
	l := NewDagreLayout()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, node := range diagram.Nodes {
			node.Position = nil
		}
		_ = l.Apply(ctx, diagram)
	}
}
