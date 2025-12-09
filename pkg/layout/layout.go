// Package layout provides diagram layout algorithms.
// This package wraps D2's built-in layout engines (Dagre, ELK) to compute
// node positions and edge routes for diagrams.
package layout

import (
	"context"
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
)

// LayoutEngine represents the type of layout algorithm to use.
type LayoutEngine string

const (
	// LayoutEngineDagre uses the Dagre hierarchical layout algorithm.
	// Good for directed graphs, flowcharts, and hierarchical diagrams.
	LayoutEngineDagre LayoutEngine = "dagre"

	// LayoutEngineELK uses the Eclipse Layout Kernel.
	// Good for complex diagrams with many connections.
	LayoutEngineELK LayoutEngine = "elk"
)

// Direction represents the primary layout direction.
type Direction string

const (
	DirectionDown  Direction = "down"  // Top to bottom (default)
	DirectionUp    Direction = "up"    // Bottom to top
	DirectionRight Direction = "right" // Left to right
	DirectionLeft  Direction = "left"  // Right to left
)

// Options configures the layout algorithm behavior.
type Options struct {
	// Engine selects the layout algorithm (default: Dagre)
	Engine LayoutEngine

	// Direction sets the primary flow direction (default: down)
	Direction Direction

	// NodeSep is the minimum separation between nodes (default: 60)
	NodeSep int

	// EdgeSep is the minimum separation between edges (default: 20)
	EdgeSep int

	// RankSep is the minimum separation between ranks/levels (default: 60)
	RankSep int

	// Padding is the padding around the diagram (default: 30)
	Padding int
}

// DefaultOptions returns the default layout options.
func DefaultOptions() Options {
	return Options{
		Engine:    LayoutEngineDagre,
		Direction: DirectionDown,
		NodeSep:   60,
		EdgeSep:   20,
		RankSep:   60,
		Padding:   30,
	}
}

// Layout is the interface for layout engines.
type Layout interface {
	// Apply computes positions for all nodes and routes for all edges.
	Apply(ctx context.Context, diagram *ir.Diagram) error
}

// DagreLayout implements layout using D2's Dagre engine.
type DagreLayout struct {
	Options Options
}

// NewDagreLayout creates a new Dagre layout engine with default options.
func NewDagreLayout() *DagreLayout {
	return &DagreLayout{
		Options: DefaultOptions(),
	}
}

// NewDagreLayoutWithOptions creates a new Dagre layout engine with custom options.
func NewDagreLayoutWithOptions(opts Options) *DagreLayout {
	return &DagreLayout{
		Options: opts,
	}
}

// Apply computes layout for the diagram using Dagre algorithm.
func (l *DagreLayout) Apply(ctx context.Context, diagram *ir.Diagram) error {
	// Convert IR back to D2 source for layout computation
	d2Source := irToD2Source(diagram, l.Options.Direction)

	// Use d2lib.Compile which handles all setup (fonts, text measurement, etc.)
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return fmt.Errorf("failed to create text ruler: %w", err)
	}

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return func(ctx context.Context, g *d2graph.Graph) error {
			dagreOpts := &d2dagrelayout.ConfigurableOpts{
				NodeSep: l.Options.NodeSep,
				EdgeSep: l.Options.EdgeSep,
			}
			return d2dagrelayout.Layout(ctx, g, dagreOpts)
		}, nil
	}

	compileOpts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}

	_, graph, err := d2lib.Compile(ctx, d2Source, compileOpts, nil)
	if err != nil {
		return fmt.Errorf("layout compilation failed: %w", err)
	}

	// Copy positions back to IR
	copyLayoutToIR(graph, diagram)

	return nil
}

// ApplyFromSource applies layout to a diagram parsed from D2 source.
// This is more efficient when you have the original D2 source.
func ApplyFromSource(ctx context.Context, source string, diagram *ir.Diagram, opts Options) error {
	// Create text ruler for measurement
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return fmt.Errorf("failed to create text ruler: %w", err)
	}

	// Create layout resolver based on engine selection
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return func(ctx context.Context, g *d2graph.Graph) error {
			switch opts.Engine {
			case LayoutEngineDagre:
				dagreOpts := &d2dagrelayout.ConfigurableOpts{
					NodeSep: opts.NodeSep,
					EdgeSep: opts.EdgeSep,
				}
				return d2dagrelayout.Layout(ctx, g, dagreOpts)
			default:
				return d2dagrelayout.DefaultLayout(ctx, g)
			}
		}, nil
	}

	compileOpts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}

	_, graph, err := d2lib.Compile(ctx, source, compileOpts, nil)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Copy positions to IR
	copyLayoutToIR(graph, diagram)

	return nil
}

// irToD2Source converts IR diagram back to D2 source for layout.
// This is needed because D2's layout engine works on its own graph structure.
func irToD2Source(diagram *ir.Diagram, direction Direction) string {
	var sb strings.Builder

	// Add direction directive
	switch direction {
	case DirectionRight:
		sb.WriteString("direction: right\n")
	case DirectionLeft:
		sb.WriteString("direction: left\n")
	case DirectionUp:
		sb.WriteString("direction: up\n")
	default:
		sb.WriteString("direction: down\n")
	}
	sb.WriteString("\n")

	// Track which nodes are containers
	containers := make(map[string]bool)
	for _, node := range diagram.Nodes {
		if node.Shape == ir.ShapeContainer {
			containers[node.ID] = true
		}
	}

	// Write root-level nodes first (those without containers)
	for _, node := range diagram.Nodes {
		if node.Container == "" && node.GetParentID() == "" {
			writeNodeToD2(&sb, node, diagram, containers, 0)
		}
	}

	sb.WriteString("\n")

	// Write edges
	for _, edge := range diagram.Edges {
		writeEdgeToD2(&sb, edge)
	}

	return sb.String()
}

// writeNodeToD2 writes a node and its children to D2 format.
func writeNodeToD2(sb *strings.Builder, node *ir.Node, diagram *ir.Diagram, containers map[string]bool, indent int) {
	prefix := strings.Repeat("  ", indent)

	// Get the local ID (last part of hierarchical ID)
	localID := node.ID
	if idx := strings.LastIndex(node.ID, "."); idx >= 0 {
		localID = node.ID[idx+1:]
	}

	// Write node declaration
	if node.Label != "" && node.Label != localID {
		sb.WriteString(fmt.Sprintf("%s%s: %s", prefix, localID, node.Label))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s", prefix, localID))
	}

	// Check if this is a container or has styling
	isContainer := containers[node.ID]
	hasStyle := node.Shape != ir.ShapeRectangle && node.Shape != ir.ShapeContainer

	if isContainer || hasStyle {
		sb.WriteString(" {\n")

		// Write shape if not default
		if hasStyle && node.Shape != ir.ShapeRectangle {
			sb.WriteString(fmt.Sprintf("%s  shape: %s\n", prefix, shapeToD2(node.Shape)))
		}

		// Write children
		if isContainer {
			children := diagram.GetNodesByContainer(node.ID)
			for _, child := range children {
				writeNodeToD2(sb, child, diagram, containers, indent+1)
			}
		}

		sb.WriteString(fmt.Sprintf("%s}\n", prefix))
	} else {
		// Add shape if not default rectangle
		if node.Shape != ir.ShapeRectangle && node.Shape != ir.ShapeContainer {
			sb.WriteString(fmt.Sprintf(" { shape: %s }\n", shapeToD2(node.Shape)))
		} else {
			sb.WriteString("\n")
		}
	}
}

// writeEdgeToD2 writes an edge in D2 format.
func writeEdgeToD2(sb *strings.Builder, edge *ir.Edge) {
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
		sb.WriteString(fmt.Sprintf("%s %s %s: %s\n", edge.Source, arrow, edge.Target, edge.Label))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s %s\n", edge.Source, arrow, edge.Target))
	}
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

// copyLayoutToIR copies computed positions from D2 graph to IR diagram.
func copyLayoutToIR(graph *d2graph.Graph, diagram *ir.Diagram) {
	// Build a map of D2 objects by their absolute ID
	objectMap := make(map[string]*d2graph.Object)
	buildObjectMap(graph.Root, "", objectMap)

	// Copy node positions
	for _, node := range diagram.Nodes {
		if obj, ok := objectMap[node.ID]; ok && obj.Box != nil {
			node.Position = &ir.Position{
				X:      obj.TopLeft.X,
				Y:      obj.TopLeft.Y,
				Source: ir.PositionSourceLayoutEngine,
			}
			node.Width = obj.Width
			node.Height = obj.Height
		}
	}

	// Copy edge routes
	edgeIndex := make(map[string]int) // Track edge indices for same source-target pairs
	for _, edge := range diagram.Edges {
		// Find matching D2 edge
		key := edge.Source + "->" + edge.Target
		idx := edgeIndex[key]
		edgeIndex[key]++

		d2Edge := findD2Edge(graph.Edges, edge.Source, edge.Target, idx)
		if d2Edge != nil && len(d2Edge.Route) > 0 {
			edge.Points = make([]ir.Point, len(d2Edge.Route))
			for i, pt := range d2Edge.Route {
				edge.Points[i] = ir.Point{X: pt.X, Y: pt.Y}
			}
		}
	}
}

// buildObjectMap recursively builds a map of D2 objects by their absolute ID.
func buildObjectMap(obj *d2graph.Object, parentID string, m map[string]*d2graph.Object) {
	if obj == nil {
		return
	}

	for _, child := range obj.ChildrenArray {
		absID := child.ID
		if parentID != "" {
			absID = parentID + "." + child.ID
		}
		m[absID] = child
		buildObjectMap(child, absID, m)
	}
}

// findD2Edge finds a D2 edge by source, target, and index.
func findD2Edge(edges []*d2graph.Edge, source, target string, index int) *d2graph.Edge {
	count := 0
	for _, e := range edges {
		srcID := e.Src.AbsID()
		dstID := e.Dst.AbsID()
		if srcID == source && dstID == target {
			if count == index {
				return e
			}
			count++
		}
	}
	return nil
}

// GetDiagramBounds calculates the bounding box of the entire diagram.
func GetDiagramBounds(diagram *ir.Diagram) (minX, minY, maxX, maxY float64) {
	if len(diagram.Nodes) == 0 {
		return 0, 0, 0, 0
	}

	minX, minY = 1e9, 1e9
	maxX, maxY = -1e9, -1e9

	for _, node := range diagram.Nodes {
		if node.Position == nil {
			continue
		}
		if node.Position.X < minX {
			minX = node.Position.X
		}
		if node.Position.Y < minY {
			minY = node.Position.Y
		}
		right := node.Position.X + node.Width
		bottom := node.Position.Y + node.Height
		if right > maxX {
			maxX = right
		}
		if bottom > maxY {
			maxY = bottom
		}
	}

	return minX, minY, maxX, maxY
}
