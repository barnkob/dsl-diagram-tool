package ir

// Diagram represents a complete diagram with all nodes and edges.
type Diagram struct {
	// Identity
	ID string `json:"id"` // Unique diagram identifier

	// Content
	Nodes []*Node `json:"nodes"` // All nodes in the diagram
	Edges []*Edge `json:"edges"` // All connections

	// Metadata
	Metadata map[string]string `json:"metadata,omitempty"` // Diagram-level metadata (title, author, etc.)

	// Configuration
	Config DiagramConfig `json:"config,omitempty"` // Rendering configuration
}

// DiagramConfig holds rendering and layout configuration.
type DiagramConfig struct {
	Theme        string `json:"theme,omitempty"`         // Theme name
	LayoutEngine string `json:"layout_engine,omitempty"` // Layout engine (dagre, elk, tala)
	Direction    string `json:"direction,omitempty"`     // Layout direction (TB, LR, etc.)
}

// GetNode returns a node by ID, or nil if not found.
func (d *Diagram) GetNode(id string) *Node {
	for _, node := range d.Nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// GetEdge returns an edge by ID, or nil if not found.
func (d *Diagram) GetEdge(id string) *Edge {
	for _, edge := range d.Edges {
		if edge.ID == id {
			return edge
		}
	}
	return nil
}

// GetNodesByContainer returns all nodes within a specific container.
func (d *Diagram) GetNodesByContainer(containerID string) []*Node {
	var nodes []*Node
	for _, node := range d.Nodes {
		if node.Container == containerID {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetEdgesByNode returns all edges connected to a specific node.
func (d *Diagram) GetEdgesByNode(nodeID string) []*Edge {
	var edges []*Edge
	for _, edge := range d.Edges {
		if edge.Source == nodeID || edge.Target == nodeID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetRootNodes returns all top-level nodes (nodes without a container).
func (d *Diagram) GetRootNodes() []*Node {
	var nodes []*Node
	for _, node := range d.Nodes {
		if node.Container == "" && node.GetParentID() == "" {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
