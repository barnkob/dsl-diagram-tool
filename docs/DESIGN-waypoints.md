# Waypoints Design Document

## Overview

This document describes the waypoints feature for the DSL Diagram Tool's browser-based editor. The design is inspired by [Structurizr's diagram editor](https://docs.structurizr.com/ui/diagrams/editor).

## Design Goals

1. Allow users to customize edge/connection routing by adding waypoints (vertices)
2. Persist waypoint data separately from the D2 source (in `.d2meta` files)
3. Provide intuitive Structurizr-like interaction patterns
4. Support different routing modes (direct, orthogonal)

## Structurizr Reference Behavior

Based on [Structurizr's documentation](https://structurizr.com/help/diagram-editor):

- **Adding vertices**: Click on an edge/link to add a waypoint
- **Moving vertices**: Drag the colored circle to reposition
- **Removing vertices**: Click the X button on the waypoint
- **Routing modes**: Press 'R' while hovering over an edge to cycle through:
  - Direct (straight lines)
  - Orthogonal (right-angle routing)
  - Curved (not implemented in our tool)

## Architecture

### Data Storage

Waypoints are stored in `.d2meta` files alongside `.d2` source files:

```json
{
  "version": 1,
  "positions": { ... },
  "waypoints": {
    "source -> target": [
      { "x": 100, "y": 200 },
      { "x": 150, "y": 250 }
    ]
  },
  "routingMode": {
    "source -> target": "orthogonal"
  },
  "sourceHash": "abc123..."
}
```

Key design decisions:
- **Separate from D2 source**: Layout metadata doesn't pollute the diagram definition
- **Source hash validation**: When the D2 source changes, all waypoints are cleared to prevent stale/invalid waypoints
- **Edge ID normalization**: Edge IDs may contain HTML entities (e.g., `-&gt;`) which are normalized for consistent storage

### Backend (Go)

Located in `pkg/server/`:

#### metadata.go

Defines the `Metadata` struct:

```go
type Metadata struct {
    Version     int                      `json:"version"`
    Positions   map[string]NodeOffset    `json:"positions"`
    Waypoints   map[string][]EdgePoint   `json:"waypoints,omitempty"`
    RoutingMode map[string]string        `json:"routingMode,omitempty"`
    SourceHash  string                   `json:"sourceHash"`
}

type EdgePoint struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}
```

Key methods:
- `SetWaypoints(edgeID, waypoints)` - Store waypoints for an edge
- `GetWaypoints(edgeID)` - Retrieve waypoints
- `SetRoutingMode(edgeID, mode)` - Store routing mode
- `GetRoutingMode(edgeID)` - Retrieve routing mode (defaults to "direct")
- `NormalizeEdgeID(edgeID)` - Decode HTML entities for consistent keys

#### handlers.go

WebSocket message handling in `handleWebSocket()`:

```go
type WSMessage struct {
    // Waypoint-related fields
    EdgeID        string                 `json:"edgeId,omitempty"`
    EdgeWaypoints []EdgePoint            `json:"waypoints,omitempty"`
    AllWaypoints  map[string][]EdgePoint `json:"allWaypoints,omitempty"`

    // Routing mode fields
    RoutingMode    string            `json:"routingMode,omitempty"`
    AllRoutingMode map[string]string `json:"allRoutingMode,omitempty"`
}
```

Message types:
- `waypoints` - Client sends to save waypoints for an edge
- `waypoints-saved` - Server acknowledgment
- `routing` - Client sends to save routing mode for an edge
- `routing-saved` - Server acknowledgment
- `positions` - Initial load includes `allWaypoints` and `allRoutingMode`
- `positions-cleared` - Broadcast when layout is reset

#### server.go

Persistence methods:
- `SetEdgeWaypoints(edgeID, waypoints)` - Update and persist
- `SetRoutingMode(edgeID, mode)` - Update and persist
- `GetMetadata()` - Returns copy with all waypoints and routing modes

### Frontend (JavaScript)

Located in `pkg/server/web/dist/index.html`:

#### State Variables

```javascript
let edgeWaypoints = {};   // { edgeId: [{ x, y, active }, ...] }
let edgeRoutingMode = {}; // { edgeId: "direct"|"orthogonal" }
let hoveredEdgeId = null; // Track for keyboard shortcuts
```

**Waypoint structure:**
- `x`, `y`: SVG coordinates
- `active`: Boolean flag - `false` for newly added waypoints, `true` after first drag
- Inactive waypoints are displayed but don't affect edge routing
- Only active waypoints are saved to server (the `active` flag is stripped before saving)

#### Key Functions

**Setup and Rendering:**
- `setupEdgeWaypoints()` - Initialize clickable edge overlays after SVG render
- `renderWaypoints()` - Draw waypoint handles with X delete buttons
- `findAllEdges()` - Locate edge groups in SVG by decoding base64 class names

**Waypoint Manipulation:**
- `addWaypoint(edgeId, x, y)` - Add waypoint, insert at optimal segment position
- `removeWaypoint(edgeId, index)` - Remove waypoint by index
- `saveWaypoints(edgeId)` - Send to server via WebSocket

**Path Calculation:**
- `calculateEdgePathWithWaypoints()` - Main entry point, dispatches to routing mode
- `calculateDirectPath()` - Straight-line segments through waypoints
- `calculateOrthogonalPath()` - Right-angle segments through waypoints
- `getRectEdgePoint()` - Calculate intersection point on node boundary
- `shortenPoint()` - Adjust endpoint for arrow marker offset

**Routing Mode:**
- `toggleRoutingMode(edgeId)` - Cycle direct -> orthogonal -> direct
- `saveRoutingMode(edgeId, mode)` - Send to server via WebSocket

## Interaction Design

### Adding Waypoints

1. Hover over an edge - shows highlighted clickable area (20px stroke width, semi-transparent blue)
2. **Double-click** on the edge - adds an **inactive** waypoint at click position
3. Inactive waypoints appear with dashed outline and don't affect edge routing yet
4. When you start dragging an inactive waypoint, it becomes active and the edge routes through it
5. Alternative: Press 'V' key while hovering to add waypoint at cursor position

**Segment insertion algorithm:**
```javascript
// Find which segment the click is closest to
for (let i = 0; i < points.length - 1; i++) {
    const dist = pointToSegmentDistance(x, y, points[i], points[i + 1]);
    if (dist < minDist) {
        minDist = dist;
        insertIndex = i;
    }
}
waypoints.splice(insertIndex, 0, { x, y });
```

### Moving Waypoints

1. Waypoints appear as blue circles (6px radius)
2. Drag the circle to reposition
3. Edge path updates in real-time during drag
4. Position is saved to server on mouse-up

### Removing Waypoints

1. Hover over waypoint - reveals red X delete button (positioned top-right)
2. Click the X button to remove waypoint
3. Edge path updates immediately

### Routing Modes

1. Hover over an edge
2. Press 'R' key to cycle through routing modes
3. Toast notification shows current mode
4. Mode is saved to server and persisted

### Keyboard Shortcuts

| Key | Action | Context |
|-----|--------|---------|
| R | Toggle routing mode | While hovering over edge |
| V | Add waypoint at cursor | While hovering over edge |

## Visual Design

```css
/* Clickable edge overlay */
.edge-clickable {
    stroke: rgba(74, 111, 243, 0);  /* Transparent until hover */
    stroke-width: 20;                /* Wide hit area */
    cursor: crosshair;
}
.edge-clickable:hover {
    stroke: rgba(74, 111, 243, 0.2); /* Semi-transparent highlight */
}

/* Waypoint handle */
.waypoint {
    fill: #4a6ff3;        /* Blue */
    stroke: #fff;
    stroke-width: 2;
    cursor: move;
}
.waypoint:hover {
    fill: #7b9df5;        /* Lighter blue */
}
.waypoint.dragging {
    fill: #2d4fb8;        /* Darker blue */
}
.waypoint.inactive {
    fill: transparent;    /* Hollow circle */
    stroke: #4a6ff3;      /* Blue outline */
    stroke-dasharray: 3 3; /* Dashed */
    opacity: 0.7;
}

/* Delete button (hidden until group hover) */
.waypoint-delete {
    fill: #e53935;        /* Red */
    opacity: 0;
}
.waypoint-group:hover .waypoint-delete {
    opacity: 1;
}
```

## Edge Path Calculation

### Edge ID Format

D2 uses base64-encoded edge IDs as CSS class names. Format after decoding:

```
(source -> target)[index]           # Top-level edge
prefix.(source -> target)[index]    # Nested edge with container prefix
```

The `parseEdgeId()` function extracts source, target, direction, and prefix.

### Direct Mode (Default)

```javascript
function calculateDirectPath(sourceBounds, targetBounds, waypoints, ...) {
    // Source point: intersection toward first waypoint
    const firstWaypoint = waypoints[0];
    let sourcePoint = getRectEdgePoint(sourceBounds, firstWaypoint.x, firstWaypoint.y);

    // Target point: intersection from last waypoint
    const lastWaypoint = waypoints[waypoints.length - 1];
    let targetPoint = getRectEdgePoint(targetBounds, lastWaypoint.x, lastWaypoint.y);

    // Build path: source -> waypoints -> target
    let path = `M ${sourcePoint.x} ${sourcePoint.y}`;
    for (const wp of waypoints) {
        path += ` L ${wp.x} ${wp.y}`;
    }
    path += ` L ${targetPoint.x} ${targetPoint.y}`;

    return path;
}
```

### Orthogonal Mode

Creates right-angle routing with bend points:

```javascript
function getOrthogonalBendPoints(from, to) {
    // Create L-shaped path using midpoints
    const midX = (from.x + to.x) / 2;
    const midY = (from.y + to.y) / 2;

    if (Math.abs(from.x - to.x) > Math.abs(from.y - to.y)) {
        // Horizontal dominant: horizontal first, then vertical
        return [{ x: midX, y: from.y }, { x: midX, y: to.y }];
    } else {
        // Vertical dominant: vertical first, then horizontal
        return [{ x: from.x, y: midY }, { x: to.x, y: midY }];
    }
}
```

### Arrow Offset

D2 arrow markers have `refX=7` with width=10, extending ~4px past path end. The path is shortened so arrow tip lands on node edge:

```javascript
const ARROW_OFFSET = 4;

function shortenPoint(point, towardX, towardY, distance) {
    const dx = towardX - point.x;
    const dy = towardY - point.y;
    const len = Math.sqrt(dx * dx + dy * dy);
    return {
        x: point.x + (dx / len) * distance,
        y: point.y + (dy / len) * distance
    };
}
```

## Known Limitations

1. **No curved routing**: Unlike Structurizr, we don't support curved/spline paths
2. **No label positioning**: Edge labels are not adjustable
3. **Self-referencing edges**: Waypoints not supported on edges where source = target
4. **Complex nesting**: Edge ID parsing may have edge cases with deeply nested containers

## Future Enhancements

1. Curved routing mode (Bezier curves through waypoints)
2. Snap-to-grid for waypoint positioning
3. Edge label position adjustment
4. Waypoint alignment guides
5. Undo/redo support for waypoint operations
6. Multi-select and bulk waypoint operations

## Testing

To test waypoints manually:

1. Start the editor: `./bin/diagtool edit examples/01-basic-shapes.d2`
2. Click on an edge to add waypoints
3. Drag waypoints to reposition
4. Press 'R' to toggle routing mode
5. Check the `.d2meta` file for persisted data
6. Reload the page to verify persistence
