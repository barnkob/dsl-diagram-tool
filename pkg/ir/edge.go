package ir

// Edge represents a connection between nodes.
type Edge struct {
	// Identity
	ID    string `json:"id"`              // Unique identifier
	Label string `json:"label,omitempty"` // Connection label

	// Connection
	Source     string    `json:"source"`                // Source node ID
	Target     string    `json:"target"`                // Target node ID
	SourcePort string    `json:"source_port,omitempty"` // Connection point on source (for SQL tables, etc.)
	TargetPort string    `json:"target_port,omitempty"` // Connection point on target
	Direction  Direction `json:"direction"`             // Arrow direction

	// Visual
	Style Style `json:"style,omitempty"` // Visual styling

	// Layout (populated by layout engine)
	Points []Point `json:"points,omitempty"` // Path coordinates

	// Extensibility
	Properties map[string]interface{} `json:"properties,omitempty"` // Custom properties
}

// Point represents a coordinate in the edge path.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// IsBidirectional returns true if the edge has arrows in both directions.
func (e *Edge) IsBidirectional() bool {
	return e.Direction == DirectionBoth
}

// HasArrowhead returns true if the edge has an arrowhead at the target.
func (e *Edge) HasArrowhead() bool {
	return e.Direction == DirectionForward || e.Direction == DirectionBoth
}

// HasArrowtail returns true if the edge has an arrowhead at the source.
func (e *Edge) HasArrowtail() bool {
	return e.Direction == DirectionBackward || e.Direction == DirectionBoth
}
