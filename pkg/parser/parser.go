// Package parser provides D2 diagram parsing capabilities.
// This package wraps the official terrastruct/d2 library and converts
// D2's internal representation to our DSL-agnostic IR.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"

	"github.com/mark/dsl-diagram-tool/pkg/ir"
)

// Parser is the interface for diagram parsers.
// Different DSL parsers (D2, PlantUML, Mermaid) implement this interface.
type Parser interface {
	// Parse converts DSL source code to internal representation.
	Parse(source string) (*ir.Diagram, error)
}

// D2Parser wraps the official terrastruct/d2 library.
type D2Parser struct {
	// Options configures parsing behavior
	Options D2ParserOptions
}

// D2ParserOptions configures the D2 parser behavior.
type D2ParserOptions struct {
	// UTF16Pos enables UTF-16 position reporting (for LSP compatibility)
	UTF16Pos bool
}

// NewD2Parser creates a new D2 parser with default options.
func NewD2Parser() *D2Parser {
	return &D2Parser{
		Options: D2ParserOptions{},
	}
}

// NewD2ParserWithOptions creates a new D2 parser with custom options.
func NewD2ParserWithOptions(opts D2ParserOptions) *D2Parser {
	return &D2Parser{
		Options: opts,
	}
}

// Parse converts D2 source code to internal representation.
func (p *D2Parser) Parse(source string) (*ir.Diagram, error) {
	// Compile D2 source to graph
	graph, _, err := d2compiler.Compile("", strings.NewReader(source), &d2compiler.CompileOptions{
		UTF16Pos: p.Options.UTF16Pos,
	})
	if err != nil {
		return nil, fmt.Errorf("d2 compilation failed: %w", err)
	}

	// Convert D2 graph to IR
	return convertGraph(graph)
}

// ParseFile reads and parses a D2 file (convenience wrapper).
func (p *D2Parser) ParseFile(source string, filename string) (*ir.Diagram, error) {
	graph, _, err := d2compiler.Compile(filename, strings.NewReader(source), &d2compiler.CompileOptions{
		UTF16Pos: p.Options.UTF16Pos,
	})
	if err != nil {
		return nil, fmt.Errorf("d2 compilation failed: %w", err)
	}

	return convertGraph(graph)
}

// convertGraph converts a D2 graph to our IR Diagram.
func convertGraph(g *d2graph.Graph) (*ir.Diagram, error) {
	diagram := &ir.Diagram{
		ID:       "diagram",
		Nodes:    make([]*ir.Node, 0),
		Edges:    make([]*ir.Edge, 0),
		Metadata: make(map[string]string),
	}

	// Convert objects to nodes (recursive for nested objects)
	if g.Root != nil {
		convertObjects(g.Root.ChildrenArray, "", diagram)
	}

	// Convert edges
	for i, edge := range g.Edges {
		irEdge := convertEdge(edge, i)
		diagram.Edges = append(diagram.Edges, irEdge)
	}

	return diagram, nil
}

// convertObjects recursively converts D2 objects to IR nodes.
func convertObjects(objects []*d2graph.Object, parentID string, diagram *ir.Diagram) {
	for _, obj := range objects {
		node := convertObject(obj, parentID)
		diagram.Nodes = append(diagram.Nodes, node)

		// Recursively convert children
		if len(obj.ChildrenArray) > 0 {
			convertObjects(obj.ChildrenArray, node.ID, diagram)
		}
	}
}

// convertObject converts a single D2 object to an IR node.
func convertObject(obj *d2graph.Object, parentID string) *ir.Node {
	// Build hierarchical ID
	id := obj.ID
	if parentID != "" {
		id = parentID + "." + obj.ID
	}

	// Determine shape type
	shape := mapD2ShapeToIR(obj)

	// Get label
	label := ""
	if obj.Label.Value != "" {
		label = obj.Label.Value
	}

	node := &ir.Node{
		ID:        id,
		Label:     label,
		Shape:     shape,
		Container: parentID,
		Style:     convertObjectStyle(obj),
	}

	// Copy position if available (from D2's layout)
	if obj.Box != nil {
		node.Position = &ir.Position{
			X:      obj.TopLeft.X,
			Y:      obj.TopLeft.Y,
			Source: ir.PositionSourceLayoutEngine,
		}
		node.Width = obj.Width
		node.Height = obj.Height
	}

	// Store D2-specific properties for extensibility
	node.Properties = make(map[string]interface{})
	if obj.Tooltip != nil && obj.Tooltip.Value != "" {
		node.Properties["tooltip"] = obj.Tooltip.Value
	}
	if obj.Link != nil && obj.Link.Value != "" {
		node.Properties["link"] = obj.Link.Value
	}
	if obj.Icon != nil {
		node.Properties["icon"] = obj.Icon.String()
	}

	return node
}

// mapD2ShapeToIR maps D2 shape strings to IR ShapeType.
func mapD2ShapeToIR(obj *d2graph.Object) ir.ShapeType {
	// Check if this is a container (has children)
	if len(obj.ChildrenArray) > 0 {
		return ir.ShapeContainer
	}

	// Get shape from D2 object
	shape := ""
	if obj.Shape.Value != "" {
		shape = obj.Shape.Value
	}

	// Map D2 shapes to IR shapes
	switch strings.ToLower(shape) {
	case "rectangle", "":
		return ir.ShapeRectangle
	case "square":
		return ir.ShapeSquare
	case "circle":
		return ir.ShapeCircle
	case "oval", "ellipse":
		return ir.ShapeOval
	case "diamond":
		return ir.ShapeDiamond
	case "parallelogram":
		return ir.ShapeParallelogram
	case "hexagon":
		return ir.ShapeHexagon
	case "person":
		return ir.ShapePerson
	case "cloud":
		return ir.ShapeCloud
	case "cylinder", "storage":
		return ir.ShapeCylinder
	case "sql_table":
		return ir.ShapeSQLTable
	case "class":
		return ir.ShapeClass
	case "code":
		return ir.ShapeCode
	case "image":
		return ir.ShapeImage
	default:
		return ir.ShapeRectangle
	}
}

// convertObjectStyle extracts style properties from a D2 object.
func convertObjectStyle(obj *d2graph.Object) ir.Style {
	style := ir.Style{}

	if obj.Style.Fill != nil && obj.Style.Fill.Value != "" {
		style.Fill = obj.Style.Fill.Value
	}
	if obj.Style.Stroke != nil && obj.Style.Stroke.Value != "" {
		style.Stroke = obj.Style.Stroke.Value
	}
	if obj.Style.StrokeWidth != nil && obj.Style.StrokeWidth.Value != "" {
		if w, err := strconv.Atoi(obj.Style.StrokeWidth.Value); err == nil {
			style.StrokeWidth = w
		}
	}
	if obj.Style.StrokeDash != nil && obj.Style.StrokeDash.Value != "" {
		if d, err := strconv.Atoi(obj.Style.StrokeDash.Value); err == nil {
			style.StrokeDash = d
		}
	}
	if obj.Style.BorderRadius != nil && obj.Style.BorderRadius.Value != "" {
		if r, err := strconv.Atoi(obj.Style.BorderRadius.Value); err == nil {
			style.BorderRadius = r
		}
	}
	if obj.Style.Opacity != nil && obj.Style.Opacity.Value != "" {
		if o, err := strconv.ParseFloat(obj.Style.Opacity.Value, 64); err == nil {
			style.Opacity = o
		}
	}
	if obj.Style.Shadow != nil && obj.Style.Shadow.Value != "" {
		style.Shadow = obj.Style.Shadow.Value == "true"
	}
	if obj.Style.ThreeDee != nil && obj.Style.ThreeDee.Value != "" {
		style.ThreeD = obj.Style.ThreeDee.Value == "true"
	}
	if obj.Style.Multiple != nil && obj.Style.Multiple.Value != "" {
		style.Multiple = obj.Style.Multiple.Value == "true"
	}
	if obj.Style.DoubleBorder != nil && obj.Style.DoubleBorder.Value != "" {
		style.DoubleBorder = obj.Style.DoubleBorder.Value == "true"
	}
	if obj.Style.Font != nil && obj.Style.Font.Value != "" {
		style.Font = obj.Style.Font.Value
	}
	if obj.Style.FontSize != nil && obj.Style.FontSize.Value != "" {
		if s, err := strconv.Atoi(obj.Style.FontSize.Value); err == nil {
			style.FontSize = s
		}
	}
	if obj.Style.FontColor != nil && obj.Style.FontColor.Value != "" {
		style.FontColor = obj.Style.FontColor.Value
	}
	if obj.Style.Bold != nil && obj.Style.Bold.Value != "" {
		style.Bold = obj.Style.Bold.Value == "true"
	}
	if obj.Style.Italic != nil && obj.Style.Italic.Value != "" {
		style.Italic = obj.Style.Italic.Value == "true"
	}
	if obj.Style.Underline != nil && obj.Style.Underline.Value != "" {
		style.Underline = obj.Style.Underline.Value == "true"
	}
	if obj.Style.TextTransform != nil && obj.Style.TextTransform.Value != "" {
		style.TextTransform = obj.Style.TextTransform.Value
	}

	return style
}

// convertEdge converts a D2 edge to an IR edge.
func convertEdge(edge *d2graph.Edge, index int) *ir.Edge {
	// Build edge ID
	srcID := edge.Src.AbsID()
	dstID := edge.Dst.AbsID()
	edgeID := fmt.Sprintf("%s-%s-%d", srcID, dstID, index)

	// Determine direction
	direction := mapD2DirectionToIR(edge.SrcArrow, edge.DstArrow)

	// Get label
	label := ""
	if edge.Label.Value != "" {
		label = edge.Label.Value
	}

	irEdge := &ir.Edge{
		ID:        edgeID,
		Label:     label,
		Source:    srcID,
		Target:    dstID,
		Direction: direction,
		Style:     convertEdgeStyle(edge),
	}

	// Handle SQL table column connections
	if edge.SrcTableColumnIndex != nil {
		irEdge.SourcePort = fmt.Sprintf("col-%d", *edge.SrcTableColumnIndex)
	}
	if edge.DstTableColumnIndex != nil {
		irEdge.TargetPort = fmt.Sprintf("col-%d", *edge.DstTableColumnIndex)
	}

	// Copy route points if available
	if len(edge.Route) > 0 {
		irEdge.Points = make([]ir.Point, len(edge.Route))
		for i, pt := range edge.Route {
			irEdge.Points[i] = ir.Point{X: pt.X, Y: pt.Y}
		}
	}

	// Store D2-specific properties
	irEdge.Properties = make(map[string]interface{})
	if edge.IsCurve {
		irEdge.Properties["curved"] = true
	}

	return irEdge
}

// mapD2DirectionToIR maps D2 arrow configuration to IR direction.
func mapD2DirectionToIR(srcArrow, dstArrow bool) ir.Direction {
	switch {
	case srcArrow && dstArrow:
		return ir.DirectionBoth
	case srcArrow && !dstArrow:
		return ir.DirectionBackward
	case !srcArrow && dstArrow:
		return ir.DirectionForward
	default:
		return ir.DirectionNone
	}
}

// convertEdgeStyle extracts style properties from a D2 edge.
func convertEdgeStyle(edge *d2graph.Edge) ir.Style {
	style := ir.Style{}

	if edge.Style.Stroke != nil && edge.Style.Stroke.Value != "" {
		style.Stroke = edge.Style.Stroke.Value
	}
	if edge.Style.StrokeWidth != nil && edge.Style.StrokeWidth.Value != "" {
		if w, err := strconv.Atoi(edge.Style.StrokeWidth.Value); err == nil {
			style.StrokeWidth = w
		}
	}
	if edge.Style.StrokeDash != nil && edge.Style.StrokeDash.Value != "" {
		if d, err := strconv.Atoi(edge.Style.StrokeDash.Value); err == nil {
			style.StrokeDash = d
		}
	}
	if edge.Style.Opacity != nil && edge.Style.Opacity.Value != "" {
		if o, err := strconv.ParseFloat(edge.Style.Opacity.Value, 64); err == nil {
			style.Opacity = o
		}
	}
	if edge.Style.Animated != nil && edge.Style.Animated.Value != "" {
		style.Animated = edge.Style.Animated.Value == "true"
	}
	if edge.Style.FontSize != nil && edge.Style.FontSize.Value != "" {
		if s, err := strconv.Atoi(edge.Style.FontSize.Value); err == nil {
			style.FontSize = s
		}
	}
	if edge.Style.FontColor != nil && edge.Style.FontColor.Value != "" {
		style.FontColor = edge.Style.FontColor.Value
	}
	if edge.Style.Bold != nil && edge.Style.Bold.Value != "" {
		style.Bold = edge.Style.Bold.Value == "true"
	}
	if edge.Style.Italic != nil && edge.Style.Italic.Value != "" {
		style.Italic = edge.Style.Italic.Value == "true"
	}

	return style
}
