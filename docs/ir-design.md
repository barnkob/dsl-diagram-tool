# Internal Representation (IR) Design - WP03

## Overview

The Internal Representation (IR) is a DSL-agnostic data model that represents diagrams in a standardized format. It serves as the bridge between:
- **Input:** DSL parsers (D2, PlantUML, Mermaid, etc.)
- **Output:** Layout engines and renderers

## Design Goals

1. **DSL Independence:** IR should represent concepts common to all diagramming languages
2. **Completeness:** Support all features needed for MVP and beyond
3. **Extensibility:** Easy to add new properties without breaking changes
4. **Type Safety:** Leverage Go's type system for correctness
5. **Simplicity:** Keep structures intuitive and easy to work with

## Core Concepts

### 1. Diagram
The top-level container representing a complete diagram.

**Properties:**
- `ID` - Unique identifier
- `Nodes` - Collection of all nodes
- `Edges` - Collection of all connections
- `Metadata` - Diagram-level metadata (title, author, etc.)
- `Config` - Rendering configuration (theme, layout engine)

### 2. Node
Represents a visual element (shape) in the diagram.

**Properties:**
- `ID` - Unique identifier (string, e.g., "server", "aws.vpc.subnet1")
- `Label` - Display text
- `Shape` - Shape type (rectangle, circle, person, etc.)
- `Style` - Visual styling
- `Container` - Parent container ID (for nesting)
- `Position` - Coordinates (set by layout engine or metadata)
- `Size` - Width and height
- `Properties` - Extensible properties map

**Node Types:**
- Basic shapes (rectangle, circle, diamond, etc.)
- Special shapes (person, cloud, cylinder)
- Containers (groups other nodes)
- SQL tables
- Classes
- Code blocks
- Images

### 3. Edge
Represents a connection between nodes.

**Properties:**
- `ID` - Unique identifier
- `Source` - Source node ID
- `Target` - Target node ID
- `SourcePort` - Connection point on source (optional, for SQL tables)
- `TargetPort` - Connection point on target (optional)
- `Label` - Connection label
- `Direction` - Arrow direction (forward, backward, both, none)
- `Style` - Visual styling
- `Points` - Path coordinates (set by layout engine)
- `Properties` - Extensible properties map

**Direction Types:**
- `Forward` (->)
- `Backward` (<-)
- `Both` (<->)
- `None` (--)

### 4. Style
Visual appearance properties.

**Shape Style Properties:**
- `Fill` - Fill color (hex, named color, gradient)
- `Stroke` - Border color
- `StrokeWidth` - Border width
- `StrokeDash` - Dash pattern
- `BorderRadius` - Corner rounding
- `Opacity` - Transparency (0.0-1.0)
- `Shadow` - Drop shadow
- `3D` - 3D effect
- `Multiple` - Multiple stacked appearance
- `DoubleBorder` - Double border

**Text Style Properties:**
- `Font` - Font family
- `FontSize` - Font size
- `FontColor` - Text color
- `Bold` - Bold text
- `Italic` - Italic text
- `Underline` - Underlined text
- `TextTransform` - Text case transformation

**Edge Style Properties:**
- `Stroke` - Line color
- `StrokeWidth` - Line width
- `StrokeDash` - Dash pattern
- `Animated` - Animation effect
- `Opacity` - Transparency

### 5. Position
Spatial coordinates for nodes.

**Properties:**
- `X` - Horizontal position
- `Y` - Vertical position
- `Width` - Element width
- `Height` - Element height
- `Source` - How position was determined (layout_engine, metadata, manual)

### 6. Container
A node that contains other nodes (composition).

Containers are represented as special nodes with:
- `Shape` = "container"
- `Children` - List of child node IDs
- Hierarchical structure using dot notation (e.g., "aws.vpc.subnet1")

## Hierarchical Structure

Nodes use hierarchical IDs with dot notation:
```
aws                    (container)
aws.vpc                (container within aws)
aws.vpc.subnet1        (container within vpc)
aws.vpc.subnet1.server (node within subnet1)
```

This enables:
- Natural parent-child relationships
- Cross-boundary connections
- Scoped styling
- Namespace organization

## Extensibility

### Properties Map
Both Node and Edge have a `Properties map[string]interface{}` field for:
- Custom attributes from specific DSLs
- Future features without schema changes
- Metadata from external tools
- User-defined data

### Type Extensions
Future node types can be added without breaking existing code:
- New Shape enum values
- Special handling in renderers
- Backward compatibility maintained

## Data Flow

```
D2 File
  ↓
D2 Parser (terrastruct/d2)
  ↓
IR Builder (our code)
  ↓
Internal Representation
  ↓
Metadata Merger (overlay metadata)
  ↓
Layout Engine (position nodes/edges)
  ↓
Renderer (SVG/PNG/PDF)
  ↓
Output File
```

## Example: Simple Diagram

**D2 Input:**
```d2
server: Web Server {
  style.fill: "#4CAF50"
}
database: Database {
  shape: cylinder
}
server -> database: SQL
```

**IR Representation:**
```go
Diagram{
  ID: "diagram-1",
  Nodes: []Node{
    {
      ID: "server",
      Label: "Web Server",
      Shape: ShapeRectangle,
      Style: Style{
        Fill: "#4CAF50",
      },
    },
    {
      ID: "database",
      Label: "Database",
      Shape: ShapeCylinder,
    },
  },
  Edges: []Edge{
    {
      ID: "server-database-0",
      Source: "server",
      Target: "database",
      Label: "SQL",
      Direction: DirectionForward,
    },
  },
}
```

## Example: Nested Containers

**D2 Input:**
```d2
aws: AWS Cloud {
  vpc: VPC {
    server: Web Server
  }
}
```

**IR Representation:**
```go
Diagram{
  Nodes: []Node{
    {
      ID: "aws",
      Label: "AWS Cloud",
      Shape: ShapeContainer,
    },
    {
      ID: "aws.vpc",
      Label: "VPC",
      Shape: ShapeContainer,
      Container: "aws",
    },
    {
      ID: "aws.vpc.server",
      Label: "Web Server",
      Shape: ShapeRectangle,
      Container: "aws.vpc",
    },
  },
}
```

## Future DSL Support

To add support for new DSLs (PlantUML, Mermaid, etc.):

1. **Create DSL-specific parser** in `pkg/parser/<dsl>/`
2. **Implement `Parser` interface:**
   ```go
   type Parser interface {
       Parse(input string) (*ir.Diagram, error)
   }
   ```
3. **Map DSL concepts to IR:**
   - DSL nodes → IR Nodes
   - DSL connections → IR Edges
   - DSL styling → IR Style
4. **Handle DSL-specific features:**
   - Store in Properties map
   - Add new IR fields if feature is common
   - Document mapping in parser

**Example Mappings:**

| DSL | Concept | IR Mapping |
|-----|---------|------------|
| D2 | `shape: person` | `Node{Shape: ShapePerson}` |
| PlantUML | `actor User` | `Node{Shape: ShapePerson, Label: "User"}` |
| Mermaid | `A[Label]` | `Node{ID: "A", Label: "Label", Shape: ShapeRectangle}` |
| D2 | `A -> B` | `Edge{Source: "A", Target: "B", Direction: Forward}` |
| PlantUML | `A --> B` | `Edge{Source: "A", Target: "B", Direction: Forward}` |
| Mermaid | `A-->B` | `Edge{Source: "A", Target: "B", Direction: Forward}` |

## Implementation Strategy

### Phase 1: Core Types (WP03) ✅
- Define IR structs in Go
- Implement basic validation
- Write unit tests

### Phase 2: D2 Integration (WP04-06)
- Create D2→IR builder
- Map D2 concepts to IR
- Handle D2-specific features

### Phase 3: Layout Integration (WP08-12)
- Add Position to IR
- Layout engine populates positions
- Support multiple layout algorithms

### Phase 4: Metadata Layer (WP27-31)
- Position overrides
- Style overrides
- Merge metadata with IR

## Validation Rules

IR should validate:
- ✅ Unique node IDs
- ✅ Unique edge IDs
- ✅ Edge source/target exist
- ✅ Container references are valid
- ✅ Hierarchical IDs match container structure
- ✅ Style values are valid (colors, ranges)
- ✅ Required fields are present

## Testing Strategy

**Unit Tests:**
- IR struct creation
- Validation logic
- Hierarchical ID parsing
- Style merging

**Integration Tests:**
- D2 → IR conversion
- IR → Renderer output
- Round-trip preservation

**Test Fixtures:**
Use examples from `examples/` and `testdata/` to verify IR can represent:
- Simple shapes
- Complex nesting
- All connection types
- All styling options
- Edge cases

---

**Design completed:** 2025-12-04
**Next step:** Implement IR types in Go (pkg/ir/)
