# Internal Representation (IR) Package

This package defines the DSL-agnostic internal representation for diagrams.

## Purpose

The IR serves as the bridge between:
- **Input parsers** (D2, PlantUML, Mermaid) that convert DSL text to IR
- **Output renderers** (SVG, PNG, PDF) that convert IR to visual output

## Core Types

### Diagram
Top-level container representing a complete diagram.

```go
diagram := &ir.Diagram{
    ID: "my-diagram",
    Nodes: []*ir.Node{...},
    Edges: []*ir.Edge{...},
}
```

### Node
Visual elements (shapes) in the diagram.

```go
node := &ir.Node{
    ID: "server",
    Label: "Web Server",
    Shape: ir.ShapeRectangle,
    Style: ir.Style{
        Fill: "#4CAF50",
        Stroke: "#2E7D32",
    },
}
```

### Edge
Connections between nodes.

```go
edge := &ir.Edge{
    ID: "conn-1",
    Source: "server",
    Target: "database",
    Direction: ir.DirectionForward,
    Label: "SQL Query",
}
```

### Style
Visual styling properties.

```go
style := ir.Style{
    Fill: "#ff0000",
    Stroke: "#000000",
    StrokeWidth: 2,
    Bold: true,
    FontSize: 14,
}
```

## Features

### Hierarchical Structure
Nodes can be nested using dot notation:

```go
nodes := []*ir.Node{
    {ID: "aws", Shape: ir.ShapeContainer},
    {ID: "aws.vpc", Container: "aws", Shape: ir.ShapeContainer},
    {ID: "aws.vpc.server", Container: "aws.vpc", Shape: ir.ShapeRectangle},
}
```

### Validation
Diagrams can be validated for structural correctness:

```go
errors := diagram.Validate()
if len(errors) > 0 {
    // Handle validation errors
}
```

### Helper Methods
Convenient methods for querying diagram structure:

```go
// Get a node by ID
node := diagram.GetNode("server")

// Get all root-level nodes
roots := diagram.GetRootNodes()

// Get all edges connected to a node
edges := diagram.GetEdgesByNode("server")

// Check if node is a container
if node.IsContainer() {
    // ...
}
```

## Usage Example

```go
package main

import "github.com/mark/dsl-diagram-tool/pkg/ir"

func main() {
    // Create diagram
    diagram := &ir.Diagram{
        ID: "example",
        Nodes: []*ir.Node{
            {
                ID: "frontend",
                Label: "Frontend",
                Shape: ir.ShapeRectangle,
                Style: ir.Style{Fill: "#4CAF50"},
            },
            {
                ID: "backend",
                Label: "Backend API",
                Shape: ir.ShapeRectangle,
                Style: ir.Style{Fill: "#2196F3"},
            },
        },
        Edges: []*ir.Edge{
            {
                ID: "fe-be",
                Source: "frontend",
                Target: "backend",
                Direction: ir.DirectionForward,
                Label: "HTTP",
            },
        },
    }

    // Validate
    if errors := diagram.Validate(); len(errors) > 0 {
        // Handle errors
    }

    // Query
    backend := diagram.GetNode("backend")
    connections := diagram.GetEdgesByNode("backend")
}
```

## Design Principles

1. **DSL Independence:** No D2-specific or PlantUML-specific concepts
2. **Extensibility:** Properties map allows custom attributes
3. **Type Safety:** Strong typing with enums for shapes, directions, etc.
4. **Simplicity:** Straightforward structs, easy to understand and use
5. **Validation:** Built-in validation for correctness

## Future Extensions

- SQL table support (ShapeSQLTable with column definitions)
- Class diagram support (ShapeClass with methods/properties)
- Sequence diagram timing information
- Grid layout metadata
- Animation specifications

See `docs/ir-design.md` for detailed design documentation.
