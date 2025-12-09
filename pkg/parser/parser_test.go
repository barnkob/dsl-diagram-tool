package parser

import (
	"testing"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
)

func TestNewD2Parser(t *testing.T) {
	p := NewD2Parser()
	if p == nil {
		t.Fatal("NewD2Parser returned nil")
	}
	if p.Options.UTF16Pos {
		t.Error("Expected UTF16Pos to be false by default")
	}
}

func TestNewD2ParserWithOptions(t *testing.T) {
	opts := D2ParserOptions{UTF16Pos: true}
	p := NewD2ParserWithOptions(opts)
	if p == nil {
		t.Fatal("NewD2ParserWithOptions returned nil")
	}
	if !p.Options.UTF16Pos {
		t.Error("Expected UTF16Pos to be true")
	}
}

func TestParse_BasicShapes(t *testing.T) {
	p := NewD2Parser()
	source := `
server
database
frontend: Frontend App
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if diagram.ID != "diagram" {
		t.Errorf("Expected diagram ID 'diagram', got '%s'", diagram.ID)
	}

	if len(diagram.Nodes) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(diagram.Nodes))
	}

	// Check nodes
	nodesByID := make(map[string]*ir.Node)
	for _, n := range diagram.Nodes {
		nodesByID[n.ID] = n
	}

	if _, ok := nodesByID["server"]; !ok {
		t.Error("Expected node 'server'")
	}
	if _, ok := nodesByID["database"]; !ok {
		t.Error("Expected node 'database'")
	}
	if n, ok := nodesByID["frontend"]; !ok {
		t.Error("Expected node 'frontend'")
	} else if n.Label != "Frontend App" {
		t.Errorf("Expected label 'Frontend App', got '%s'", n.Label)
	}
}

func TestParse_Connections(t *testing.T) {
	p := NewD2Parser()
	source := `
a -> b: forward
c <- d: backward
e <-> f: bidirectional
g -- h: no arrow
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Edges) != 4 {
		t.Fatalf("Expected 4 edges, got %d", len(diagram.Edges))
	}

	// Build edge map by checking source-target pairs
	edgeMap := make(map[string]*ir.Edge)
	for _, e := range diagram.Edges {
		key := e.Source + "->" + e.Target
		edgeMap[key] = e
	}

	// Forward arrow: a -> b
	if e, ok := edgeMap["a->b"]; !ok {
		t.Error("Expected edge a->b")
	} else {
		if e.Direction != ir.DirectionForward {
			t.Errorf("Expected forward direction, got %s", e.Direction)
		}
		if e.Label != "forward" {
			t.Errorf("Expected label 'forward', got '%s'", e.Label)
		}
	}

	// Backward arrow: c <- d (stored as d->c with backward direction, or c->d depending on D2)
	// D2 normalizes this - let's check what we get
	foundBackward := false
	for _, e := range diagram.Edges {
		if e.Label == "backward" {
			foundBackward = true
			if e.Direction != ir.DirectionBackward {
				t.Errorf("Expected backward direction for 'backward' edge, got %s", e.Direction)
			}
		}
	}
	if !foundBackward {
		t.Error("Expected to find edge with label 'backward'")
	}

	// Bidirectional: e <-> f
	foundBidi := false
	for _, e := range diagram.Edges {
		if e.Label == "bidirectional" {
			foundBidi = true
			if e.Direction != ir.DirectionBoth {
				t.Errorf("Expected both direction, got %s", e.Direction)
			}
		}
	}
	if !foundBidi {
		t.Error("Expected to find bidirectional edge")
	}

	// No arrow: g -- h
	foundNone := false
	for _, e := range diagram.Edges {
		if e.Label == "no arrow" {
			foundNone = true
			if e.Direction != ir.DirectionNone {
				t.Errorf("Expected none direction, got %s", e.Direction)
			}
		}
	}
	if !foundNone {
		t.Error("Expected to find edge with no arrow")
	}
}

func TestParse_ShapeTypes(t *testing.T) {
	p := NewD2Parser()
	source := `
rect: Rectangle
person: Person { shape: person }
cloud: Cloud { shape: cloud }
cylinder: Database { shape: cylinder }
circle: Circle { shape: circle }
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	nodesByID := make(map[string]*ir.Node)
	for _, n := range diagram.Nodes {
		nodesByID[n.ID] = n
	}

	tests := []struct {
		id    string
		shape ir.ShapeType
	}{
		{"rect", ir.ShapeRectangle},
		{"person", ir.ShapePerson},
		{"cloud", ir.ShapeCloud},
		{"cylinder", ir.ShapeCylinder},
		{"circle", ir.ShapeCircle},
	}

	for _, tt := range tests {
		n, ok := nodesByID[tt.id]
		if !ok {
			t.Errorf("Expected node '%s'", tt.id)
			continue
		}
		if n.Shape != tt.shape {
			t.Errorf("Node '%s': expected shape %s, got %s", tt.id, tt.shape, n.Shape)
		}
	}
}

func TestParse_Containers(t *testing.T) {
	p := NewD2Parser()
	source := `
aws: AWS Cloud {
  vpc: VPC {
    server: Web Server
  }
}
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(diagram.Nodes))
	}

	nodesByID := make(map[string]*ir.Node)
	for _, n := range diagram.Nodes {
		nodesByID[n.ID] = n
	}

	// Check aws
	aws, ok := nodesByID["aws"]
	if !ok {
		t.Fatal("Expected node 'aws'")
	}
	if aws.Shape != ir.ShapeContainer {
		t.Errorf("Expected aws to be container, got %s", aws.Shape)
	}
	if aws.Container != "" {
		t.Errorf("Expected aws.Container to be empty, got '%s'", aws.Container)
	}

	// Check aws.vpc
	vpc, ok := nodesByID["aws.vpc"]
	if !ok {
		t.Fatal("Expected node 'aws.vpc'")
	}
	if vpc.Shape != ir.ShapeContainer {
		t.Errorf("Expected vpc to be container, got %s", vpc.Shape)
	}
	if vpc.Container != "aws" {
		t.Errorf("Expected vpc.Container to be 'aws', got '%s'", vpc.Container)
	}

	// Check aws.vpc.server
	server, ok := nodesByID["aws.vpc.server"]
	if !ok {
		t.Fatal("Expected node 'aws.vpc.server'")
	}
	if server.Shape != ir.ShapeRectangle {
		t.Errorf("Expected server to be rectangle, got %s", server.Shape)
	}
	if server.Container != "aws.vpc" {
		t.Errorf("Expected server.Container to be 'aws.vpc', got '%s'", server.Container)
	}
	if server.Label != "Web Server" {
		t.Errorf("Expected server.Label to be 'Web Server', got '%s'", server.Label)
	}
}

func TestParse_Styling(t *testing.T) {
	p := NewD2Parser()
	source := `
styled: Styled Node {
  style: {
    fill: "#ff0000"
    stroke: "#000000"
    stroke-width: 3
    border-radius: 8
    opacity: 0.8
    shadow: true
    bold: true
    font-size: 16
  }
}
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(diagram.Nodes))
	}

	node := diagram.Nodes[0]
	style := node.Style

	if style.Fill != "#ff0000" {
		t.Errorf("Expected fill '#ff0000', got '%s'", style.Fill)
	}
	if style.Stroke != "#000000" {
		t.Errorf("Expected stroke '#000000', got '%s'", style.Stroke)
	}
	if style.StrokeWidth != 3 {
		t.Errorf("Expected stroke-width 3, got %d", style.StrokeWidth)
	}
	if style.BorderRadius != 8 {
		t.Errorf("Expected border-radius 8, got %d", style.BorderRadius)
	}
	if style.Opacity != 0.8 {
		t.Errorf("Expected opacity 0.8, got %f", style.Opacity)
	}
	if !style.Shadow {
		t.Error("Expected shadow to be true")
	}
	if !style.Bold {
		t.Error("Expected bold to be true")
	}
	if style.FontSize != 16 {
		t.Errorf("Expected font-size 16, got %d", style.FontSize)
	}
}

func TestParse_EdgeStyling(t *testing.T) {
	p := NewD2Parser()
	source := `
a -> b: styled {
  style: {
    stroke: red
    stroke-width: 2
    stroke-dash: 5
    animated: true
  }
}
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(diagram.Edges))
	}

	edge := diagram.Edges[0]
	style := edge.Style

	if style.Stroke != "red" {
		t.Errorf("Expected stroke 'red', got '%s'", style.Stroke)
	}
	if style.StrokeWidth != 2 {
		t.Errorf("Expected stroke-width 2, got %d", style.StrokeWidth)
	}
	if style.StrokeDash != 5 {
		t.Errorf("Expected stroke-dash 5, got %d", style.StrokeDash)
	}
	if !style.Animated {
		t.Error("Expected animated to be true")
	}
}

func TestParse_CrossContainerEdges(t *testing.T) {
	p := NewD2Parser()
	source := `
aws: AWS {
  server: Server
}
client: Client

client -> aws.server: API Call
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Edges) != 1 {
		t.Fatalf("Expected 1 edge, got %d", len(diagram.Edges))
	}

	edge := diagram.Edges[0]
	if edge.Source != "client" {
		t.Errorf("Expected source 'client', got '%s'", edge.Source)
	}
	if edge.Target != "aws.server" {
		t.Errorf("Expected target 'aws.server', got '%s'", edge.Target)
	}
	if edge.Label != "API Call" {
		t.Errorf("Expected label 'API Call', got '%s'", edge.Label)
	}
}

func TestParse_InvalidSyntax(t *testing.T) {
	p := NewD2Parser()
	source := `
invalid syntax {{{{
`
	_, err := p.Parse(source)
	if err == nil {
		t.Error("Expected parse error for invalid syntax")
	}
}

func TestParse_EmptySource(t *testing.T) {
	p := NewD2Parser()
	diagram, err := p.Parse("")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(diagram.Edges))
	}
}

func TestParse_Comments(t *testing.T) {
	p := NewD2Parser()
	source := `
# This is a comment
server
# Another comment
database
server -> database # Inline comment
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(diagram.Edges))
	}
}

func TestParse_MultipleEdgesBetweenSameNodes(t *testing.T) {
	p := NewD2Parser()
	source := `
a -> b: first
a -> b: second
a -> b: third
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Edges) != 3 {
		t.Fatalf("Expected 3 edges, got %d", len(diagram.Edges))
	}

	// Check that all edges have unique IDs
	edgeIDs := make(map[string]bool)
	for _, e := range diagram.Edges {
		if edgeIDs[e.ID] {
			t.Errorf("Duplicate edge ID: %s", e.ID)
		}
		edgeIDs[e.ID] = true
	}
}

func TestMapD2DirectionToIR(t *testing.T) {
	tests := []struct {
		srcArrow bool
		dstArrow bool
		expected ir.Direction
	}{
		{false, true, ir.DirectionForward},
		{true, false, ir.DirectionBackward},
		{true, true, ir.DirectionBoth},
		{false, false, ir.DirectionNone},
	}

	for _, tt := range tests {
		result := mapD2DirectionToIR(tt.srcArrow, tt.dstArrow)
		if result != tt.expected {
			t.Errorf("mapD2DirectionToIR(%v, %v) = %s, expected %s",
				tt.srcArrow, tt.dstArrow, result, tt.expected)
		}
	}
}

func TestParse_NodeProperties(t *testing.T) {
	p := NewD2Parser()
	source := `
linked: Linked Node {
  link: https://example.com
  tooltip: This is a tooltip
}
`
	diagram, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(diagram.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(diagram.Nodes))
	}

	node := diagram.Nodes[0]
	if node.Properties == nil {
		t.Fatal("Expected node.Properties to be non-nil")
	}

	if link, ok := node.Properties["link"]; !ok {
		t.Error("Expected 'link' property")
	} else if link != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got '%v'", link)
	}

	if tooltip, ok := node.Properties["tooltip"]; !ok {
		t.Error("Expected 'tooltip' property")
	} else if tooltip != "This is a tooltip" {
		t.Errorf("Expected tooltip 'This is a tooltip', got '%v'", tooltip)
	}
}

func TestParseFile(t *testing.T) {
	p := NewD2Parser()
	source := `server -> database`
	diagram, err := p.ParseFile(source, "test.d2")
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(diagram.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(diagram.Nodes))
	}
	if len(diagram.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(diagram.Edges))
	}
}

// Benchmark tests
func BenchmarkParse_Simple(b *testing.B) {
	p := NewD2Parser()
	source := `
server -> database
database -> cache
cache -> server
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_Complex(b *testing.B) {
	p := NewD2Parser()
	source := `
aws: AWS Cloud {
  vpc: VPC {
    public: Public Subnet {
      lb: Load Balancer
      web1: Web Server 1
      web2: Web Server 2
    }
    private: Private Subnet {
      app1: App Server 1
      app2: App Server 2
      db: Database {
        shape: cylinder
      }
    }
  }
}

lb -> web1
lb -> web2
web1 -> app1
web2 -> app2
app1 -> db
app2 -> db
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}
