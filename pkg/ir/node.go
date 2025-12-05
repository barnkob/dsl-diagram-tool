package ir

// Node represents a visual element (shape) in the diagram.
type Node struct {
	// Identity
	ID    string `json:"id"`              // Unique identifier (hierarchical, e.g., "aws.vpc.subnet1")
	Label string `json:"label,omitempty"` // Display text

	// Type
	Shape ShapeType `json:"shape"` // Shape type

	// Hierarchy
	Container string `json:"container,omitempty"` // Parent container ID

	// Visual
	Style Style `json:"style,omitempty"` // Visual styling

	// Layout (populated by layout engine)
	Position *Position `json:"position,omitempty"` // Spatial position
	Width    float64   `json:"width,omitempty"`    // Element width
	Height   float64   `json:"height,omitempty"`   // Element height

	// Extensibility
	Properties map[string]interface{} `json:"properties,omitempty"` // Custom properties
}

// Position represents the spatial coordinates of a node.
type Position struct {
	X      float64        `json:"x"`      // Horizontal position
	Y      float64        `json:"y"`      // Vertical position
	Source PositionSource `json:"source"` // How position was determined
}

// IsContainer returns true if this node is a container.
func (n *Node) IsContainer() bool {
	return n.Shape == ShapeContainer
}

// GetHierarchyLevel returns the nesting level (0 for root, 1 for first level, etc.).
func (n *Node) GetHierarchyLevel() int {
	if n.ID == "" {
		return 0
	}

	level := 0
	for _, ch := range n.ID {
		if ch == '.' {
			level++
		}
	}
	return level
}

// GetParentID returns the parent container ID from the hierarchical ID.
// Returns empty string if this is a root node.
func (n *Node) GetParentID() string {
	if n.Container != "" {
		return n.Container
	}

	// Parse from hierarchical ID
	for i := len(n.ID) - 1; i >= 0; i-- {
		if n.ID[i] == '.' {
			return n.ID[:i]
		}
	}
	return ""
}
