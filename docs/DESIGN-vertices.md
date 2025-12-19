# Vertices Design Document

## Overview

This document describes the vertices (edge bend points) feature for the DSL Diagram Tool's browser-based editor. The design is inspired by [Structurizr's diagram editor](https://docs.structurizr.com/ui/diagrams/editor).

## Design Goals

1. Allow users to customize edge/connection routing by adding vertices (bend points)
2. Persist vertex data separately from the D2 source (in `.d2meta` files)
3. Provide intuitive interaction patterns using JointJS
4. Support drag-and-drop for both nodes and edge vertices

## Architecture

### Frontend (JointJS)

The browser editor uses [JointJS](https://www.jointjs.com/) for diagram rendering and interaction.

**Key Benefits:**
- Built-in drag-and-drop for nodes
- Built-in vertex manipulation for edges via `linkTools.Vertices`
- Professional diagramming library with extensive features
- Handles coordinate transformations, hit detection, and rendering

**JointJS Components Used:**
- `joint.dia.Graph` - Data model for diagram elements
- `joint.dia.Paper` - SVG rendering surface
- `joint.shapes.standard.Rectangle` - Node shapes
- `joint.shapes.standard.Link` - Edge connections
- `joint.linkTools.Vertices` - Interactive vertex tools

### Data Flow

1. D2 source is rendered to SVG on the server
2. Frontend parses SVG to extract node positions/sizes and edge connections
3. JointJS elements are created from parsed data
4. Stored positions and vertices from `.d2meta` are applied
5. User interactions update JointJS model
6. Changes are synced to server via WebSocket

### Data Storage

Vertices are stored in `.d2meta` files alongside `.d2` source files:

```json
{
  "version": 1,
  "positions": {
    "nodeId": { "dx": 50, "dy": -30 }
  },
  "vertices": {
    "source -> target": [
      { "x": 100, "y": 200 },
      { "x": 150, "y": 250 }
    ]
  },
  "sourceHash": "abc123..."
}
```

Key design decisions:
- **Separate from D2 source**: Layout metadata doesn't pollute the diagram definition
- **Source hash validation**: When the D2 source changes, all vertices are cleared to prevent stale/invalid data
- **Edge ID normalization**: Edge IDs may contain HTML entities (e.g., `-&gt;`) which are normalized for consistent storage

### Backend (Go)

Located in `pkg/server/`:

#### metadata.go

Defines the `Metadata` struct:

```go
type Metadata struct {
    Version     int                   `json:"version"`
    Positions   map[string]NodeOffset `json:"positions"`
    Vertices    map[string][]Vertex   `json:"vertices,omitempty"`
    RoutingMode map[string]string     `json:"routingMode,omitempty"`
    SourceHash  string                `json:"sourceHash"`
}

type Vertex struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}
```

Key methods:
- `SetVertices(edgeID, vertices)` - Store vertices for an edge
- `GetVertices(edgeID)` - Retrieve vertices
- `NormalizeEdgeID(edgeID)` - Decode HTML entities for consistent keys

#### handlers.go

WebSocket message handling:

```go
type WSMessage struct {
    // Vertex-related fields
    EdgeID      string              `json:"edgeId,omitempty"`
    Vertices    []Vertex            `json:"vertices,omitempty"`
    AllVertices map[string][]Vertex `json:"allVertices,omitempty"`
}
```

Message types:
- `vertices` - Client sends to save vertices for an edge
- `vertices-saved` - Server acknowledgment
- `positions` - Initial load includes `allVertices`
- `positions-cleared` - Broadcast when layout is reset

### Frontend (JavaScript)

Located in `pkg/server/web/dist/index.html`:

#### State Variables

```javascript
let graph;              // JointJS graph model
let paper;              // JointJS paper (SVG canvas)
let jointElements = {}; // Map D2 nodeId -> JointJS element
let jointLinks = {};    // Map D2 edgeId -> JointJS link
let nodePositions = {}; // Metadata positions from server
let edgeVertices = {};  // Metadata vertices from server
```

#### Key Event Handlers

**Vertex Changes:**
```javascript
graph.on('change:vertices', (link, vertices) => {
    const edgeId = link.get('edgeId');
    // Update local state
    if (vertices && vertices.length > 0) {
        edgeVertices[edgeId] = vertices;
    } else {
        delete edgeVertices[edgeId];
    }
    // Send to server
    ws.send(JSON.stringify({
        type: 'vertices',
        edgeId: edgeId,
        vertices: vertices || []
    }));
});
```

**Link Selection:**
```javascript
paper.on('link:pointerclick', (linkView) => {
    // Add vertex tools to clicked link
    const tools = new joint.dia.ToolsView({
        tools: [
            new joint.linkTools.Vertices({
                snapRadius: 10,
                redundancyRemoval: false,
                vertexAdding: true,
                vertexRemoving: true  // Double-click to remove
            })
        ]
    });
    linkView.addTools(tools);
});
```

## Interaction Design

### Adding Vertices

1. Click on an edge to select it - shows vertex tool handles
2. Click anywhere on the edge path to add a new vertex
3. The vertex appears as a draggable circle

### Moving Vertices

1. Vertices appear as blue circles when edge is selected
2. Drag the circle to reposition
3. Edge path updates in real-time during drag
4. Position is saved to server on mouse-up

### Removing Vertices

1. Select the edge by clicking on it
2. Double-click on a vertex to remove it
3. Edge path updates immediately
4. Change is saved to server

### Deselecting

1. Click on empty canvas area to deselect edge
2. Vertex tools are hidden

## Visual Design

JointJS provides default styling for vertex handles. Custom CSS can be applied:

```css
/* Vertex handle styling */
.joint-link .marker-vertex {
    fill: #4a6ff3;
    stroke: #fff;
    stroke-width: 2;
    r: 6;
}
.joint-link .marker-vertex:hover {
    fill: #7b9df5;
    cursor: move;
}
```

## Known Limitations

1. **No routing modes**: Unlike previous implementation, orthogonal routing is not yet implemented with JointJS
2. **No keyboard shortcuts**: 'R' and 'V' shortcuts not implemented
3. **Self-referencing edges**: Vertices may not work on edges where source = target
4. **Complex nesting**: Edge ID parsing may have edge cases with deeply nested containers

## Future Enhancements

1. Orthogonal routing mode via JointJS routers
2. Snap-to-grid for vertex positioning
3. Keyboard shortcuts for quick vertex operations
4. Edge label position adjustment
5. Undo/redo support

## Testing

To test vertices manually:

1. Start the editor: `./bin/diagtool serve examples/01-basic-shapes.d2`
2. Click on an edge to select it (shows vertex handles)
3. Click on edge path to add vertices
4. Drag vertices to reposition
5. Double-click vertices to remove them
6. Check the `.d2meta` file for persisted data
7. Reload the page to verify persistence
