# Layout Metadata Design Document

## Overview

This document captures research on how diagramming tools handle layout metadata separately from source content, with a focus on Structurizr's approach and recommendations for the DSL Diagram Tool.

## Structurizr Reference Architecture

Based on research from:
- [Structurizr Manual Layout](https://docs.structurizr.com/ui/diagrams/manual-layout)
- [Structurizr Diagram Editor](https://docs.structurizr.com/ui/diagrams/editor)
- [Structurizr JSON Schema](https://github.com/structurizr/json)

### Core Principle

Structurizr separates **model content** (DSL) from **layout metadata** (JSON):

> "Diagram layout information isn't something that you will ever author by hand. For example, element x,y positions are not stored in the source of your workspace, such as your workspace definition when using the Structurizr DSL. This information is instead stored in a JSON version of your workspace."

### Structurizr JSON Schema

**ElementView** - Node positions:
```yaml
ElementView:
  properties:
    id: string      # Element identifier
    x: integer      # Absolute X position (pixels)
    y: integer      # Absolute Y position (pixels)
```

**RelationshipView** - Edge routing:
```yaml
RelationshipView:
  properties:
    id: string
    vertices: array of Vertex  # Waypoints/bend points
    routing: enum              # Direct, Curved, Orthogonal
    position: integer          # Label position (0-100 along line)
```

**Vertex** - Waypoint coordinates:
```yaml
Vertex:
  properties:
    x: integer
    y: integer
```

**AutomaticLayout** - Layout engine settings:
```yaml
AutomaticLayout:
  properties:
    implementation: enum       # Graphviz, Dagre
    rankDirection: enum        # TopBottom, BottomTop, LeftRight, RightLeft
    rankSeparation: integer
    nodeSeparation: integer
    edgeSeparation: integer
    vertices: boolean          # Create vertices during layout
```

### Change Handling (Merge Strategy)

When the DSL changes, Structurizr uses a **smart merging algorithm**:

1. Parse new DSL into in-memory workspace model
2. Load existing workspace JSON with layout information
3. Match elements between old and new versions:
   - Primary matching: by element name
   - Fallback matching: by internal ID
4. Preserve layout for matched elements
5. Discard layout for removed elements

**Known limitations:**
- Renaming elements while changing creation order can break matching
- Changing a view's key causes entire diagram to lose layout
- Recommendation: use explicit view keys (e.g., `systemLandscape "MyViewKey"`)

## Current Implementation (.d2meta)

### File Structure

```json
{
  "version": 1,
  "positions": {
    "nodeId": { "dx": 10, "dy": 20 }
  },
  "waypoints": {
    "source -> target": [{ "x": 100, "y": 200 }]
  },
  "routingMode": {
    "source -> target": "orthogonal"
  },
  "sourceHash": "abc123..."
}
```

### Key Differences from Structurizr

| Aspect | Structurizr | Current .d2meta |
|--------|-------------|-----------------|
| Position type | Absolute (x, y) | **Offset** (dx, dy) from auto-layout |
| Change handling | Smart merge by ID | **Clear all** (source hash) |
| Waypoints | `vertices: [{x,y}]` | `waypoints: [{x,y}]` ✓ |
| Routing modes | Direct, Curved, Orthogonal | Direct, Orthogonal (no curved) |
| Multi-view | Yes (per-view layouts) | No (single diagram) |
| Label positioning | Yes (0-100 along edge) | No |

### Pros of Current Offset Approach

1. **Works with D2 auto-layout**: Positions are deltas from computed layout
2. **Minimal storage**: Only stores differences, not all positions
3. **Graceful degradation**: If metadata is lost, diagram still renders correctly

### Cons of Current Approach

1. **Source hash clearing**: Any D2 change clears ALL layout metadata
2. **No element matching**: Can't preserve layout when adding new nodes
3. **Offset fragility**: If D2's auto-layout algorithm changes, offsets may look wrong

## Recommended Improvements

### Phase 1: Smart Merging (Priority)

Replace source hash clearing with element-based matching:

```go
// Pseudo-code for merge strategy
func (m *Metadata) SmartMerge(oldSource, newSource string) {
    oldNodes := parseNodeIDs(oldSource)
    newNodes := parseNodeIDs(newSource)

    // Preserve positions for nodes that still exist
    for nodeID, position := range m.Positions {
        if !newNodes.Contains(nodeID) {
            delete(m.Positions, nodeID)
        }
    }

    // Similar for waypoints - preserve if both endpoints exist
    for edgeID, waypoints := range m.Waypoints {
        source, target := parseEdgeEndpoints(edgeID)
        if !newNodes.Contains(source) || !newNodes.Contains(target) {
            delete(m.Waypoints, edgeID)
        }
    }
}
```

**Benefits:**
- Adding a new node preserves existing layout
- Only removes layout for deleted elements
- Much better UX for iterative diagram development

### Phase 2: Consider Absolute Positions (Optional)

Switch from offset-based to absolute positioning:

```json
{
  "positions": {
    "nodeId": { "x": 150, "y": 200 }
  }
}
```

**Trade-offs:**
- Pro: Independent of auto-layout changes
- Pro: Matches Structurizr's approach
- Con: Must store position for every manually-placed node
- Con: Loses "enhancement" feel (currently enhances auto-layout)

### Phase 3: Multi-View Support (Future)

If D2 adds multi-view support, extend metadata:

```json
{
  "views": {
    "system-context": {
      "positions": { ... },
      "waypoints": { ... }
    },
    "container": {
      "positions": { ... },
      "waypoints": { ... }
    }
  }
}
```

## JointJS Integration Assessment

### Architecture Challenge

Current flow:
```
D2 Source → D2 Library → SVG → DOM manipulation → Interactivity
```

With JointJS:
```
D2 Source → D2 Library → SVG → Parse to JointJS → Interactive editing
                                      ↓
                               Save to .d2meta
                                      ↓
                               Re-render D2 → Re-sync JointJS
```

### JointJS Pros

- Built-in drag-drop, waypoints, routing
- Undo/redo support
- Multi-select and alignment tools
- Professional diagramming UX
- Active maintenance and documentation

### JointJS Cons

- Must sync two models (D2 output ↔ JointJS)
- Complexity of parsing D2's SVG into JointJS shapes
- Potential visual mismatches between D2 and JointJS rendering
- Additional ~300KB dependency
- Learning curve for JointJS API

### Recommendation

**Start with vanilla approach enhancements:**
1. Implement smart merging (high value, low risk)
2. Add undo/redo with command pattern
3. Add multi-select with shift+click

**Then evaluate JointJS via PoC:**
1. Create proof-of-concept on separate branch
2. Test SVG parsing and model sync
3. Evaluate UX improvement vs complexity cost
4. Make data-driven decision

## References

- [Structurizr Manual Layout Documentation](https://docs.structurizr.com/ui/diagrams/manual-layout)
- [Structurizr Diagram Editor](https://docs.structurizr.com/ui/diagrams/editor)
- [Structurizr JSON Schema (GitHub)](https://github.com/structurizr/json)
- [Structurizr YAML Schema Definition](https://github.com/structurizr/json/blob/master/structurizr.yaml)
- [JointJS Documentation](https://www.jointjs.com/docs)
