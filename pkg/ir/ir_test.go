package ir

import (
	"testing"
)

func TestDiagram_GetNode(t *testing.T) {
	diagram := &Diagram{
		Nodes: []*Node{
			{ID: "node1", Label: "Node 1"},
			{ID: "node2", Label: "Node 2"},
		},
	}

	tests := []struct {
		name   string
		id     string
		expect bool
	}{
		{"existing node", "node1", true},
		{"another existing node", "node2", true},
		{"non-existent node", "node3", false},
		{"empty id", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := diagram.GetNode(tt.id)
			if (node != nil) != tt.expect {
				t.Errorf("GetNode(%q) = %v, expected found=%v", tt.id, node, tt.expect)
			}
		})
	}
}

func TestDiagram_GetRootNodes(t *testing.T) {
	diagram := &Diagram{
		Nodes: []*Node{
			{ID: "root1"},
			{ID: "root2"},
			{ID: "container.child", Container: "container"},
			{ID: "container"},
		},
	}

	roots := diagram.GetRootNodes()
	if len(roots) != 3 { // root1, root2, container
		t.Errorf("expected 3 root nodes, got %d", len(roots))
	}
}

func TestNode_IsContainer(t *testing.T) {
	tests := []struct {
		name   string
		shape  ShapeType
		expect bool
	}{
		{"container shape", ShapeContainer, true},
		{"rectangle shape", ShapeRectangle, false},
		{"circle shape", ShapeCircle, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{Shape: tt.shape}
			if node.IsContainer() != tt.expect {
				t.Errorf("IsContainer() = %v, expected %v", node.IsContainer(), tt.expect)
			}
		})
	}
}

func TestNode_GetHierarchyLevel(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		expect int
	}{
		{"root level", "node", 0},
		{"first level", "parent.child", 1},
		{"second level", "grandparent.parent.child", 2},
		{"third level", "a.b.c.d", 3},
		{"empty id", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{ID: tt.id}
			if level := node.GetHierarchyLevel(); level != tt.expect {
				t.Errorf("GetHierarchyLevel() = %d, expected %d", level, tt.expect)
			}
		})
	}
}

func TestNode_GetParentID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		container string
		expect    string
	}{
		{"explicit container", "parent.child", "parent", "parent"},
		{"hierarchical id", "parent.child", "", "parent"},
		{"deep hierarchy", "a.b.c", "", "a.b"},
		{"root node", "root", "", ""},
		{"empty id", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Node{ID: tt.id, Container: tt.container}
			if parent := node.GetParentID(); parent != tt.expect {
				t.Errorf("GetParentID() = %q, expected %q", parent, tt.expect)
			}
		})
	}
}

func TestEdge_IsBidirectional(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		expect    bool
	}{
		{"bidirectional", DirectionBoth, true},
		{"forward only", DirectionForward, false},
		{"backward only", DirectionBackward, false},
		{"no direction", DirectionNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edge := &Edge{Direction: tt.direction}
			if edge.IsBidirectional() != tt.expect {
				t.Errorf("IsBidirectional() = %v, expected %v", edge.IsBidirectional(), tt.expect)
			}
		})
	}
}

func TestEdge_HasArrowhead(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		expect    bool
	}{
		{"forward", DirectionForward, true},
		{"both", DirectionBoth, true},
		{"backward", DirectionBackward, false},
		{"none", DirectionNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edge := &Edge{Direction: tt.direction}
			if edge.HasArrowhead() != tt.expect {
				t.Errorf("HasArrowhead() = %v, expected %v", edge.HasArrowhead(), tt.expect)
			}
		})
	}
}

func TestStyle_Merge(t *testing.T) {
	base := Style{
		Fill:        "#ff0000",
		Stroke:      "#000000",
		StrokeWidth: 1,
		Bold:        true,
	}

	override := Style{
		Fill:     "#00ff00",
		FontSize: 14,
	}

	result := base.Merge(override)

	if result.Fill != "#00ff00" {
		t.Errorf("expected fill to be overridden to #00ff00, got %s", result.Fill)
	}
	if result.Stroke != "#000000" {
		t.Errorf("expected stroke to remain #000000, got %s", result.Stroke)
	}
	if result.FontSize != 14 {
		t.Errorf("expected font size to be 14, got %d", result.FontSize)
	}
	if !result.Bold {
		t.Errorf("expected bold to remain true")
	}
}

func TestDiagram_Validate(t *testing.T) {
	tests := []struct {
		name      string
		diagram   *Diagram
		expectErr bool
		errCount  int
	}{
		{
			name: "valid diagram",
			diagram: &Diagram{
				ID: "test",
				Nodes: []*Node{
					{ID: "a", Shape: ShapeRectangle},
					{ID: "b", Shape: ShapeCircle},
				},
				Edges: []*Edge{
					{ID: "e1", Source: "a", Target: "b", Direction: DirectionForward},
				},
			},
			expectErr: false,
		},
		{
			name: "duplicate node IDs",
			diagram: &Diagram{
				Nodes: []*Node{
					{ID: "a", Shape: ShapeRectangle},
					{ID: "a", Shape: ShapeCircle},
				},
			},
			expectErr: true,
			errCount:  1,
		},
		{
			name: "edge references non-existent node",
			diagram: &Diagram{
				Nodes: []*Node{
					{ID: "a", Shape: ShapeRectangle},
				},
				Edges: []*Edge{
					{ID: "e1", Source: "a", Target: "nonexistent", Direction: DirectionForward},
				},
			},
			expectErr: true,
			errCount:  1,
		},
		{
			name: "invalid container reference",
			diagram: &Diagram{
				Nodes: []*Node{
					{ID: "a", Shape: ShapeRectangle, Container: "nonexistent"},
				},
			},
			expectErr: true,
			errCount:  2, // Non-existent container + hierarchical ID mismatch
		},
		{
			name: "invalid opacity",
			diagram: &Diagram{
				Nodes: []*Node{
					{ID: "a", Shape: ShapeRectangle, Style: Style{Opacity: 1.5}},
				},
			},
			expectErr: true,
			errCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.diagram.Validate()
			if tt.expectErr && len(errors) == 0 {
				t.Error("expected validation errors but got none")
			}
			if !tt.expectErr && len(errors) > 0 {
				t.Errorf("expected no errors but got: %v", errors)
			}
			if tt.errCount > 0 && len(errors) != tt.errCount {
				t.Errorf("expected %d errors but got %d: %v", tt.errCount, len(errors), errors)
			}
		})
	}
}
