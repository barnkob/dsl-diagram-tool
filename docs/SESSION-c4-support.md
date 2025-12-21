# Session: C4 Model Support

**Date:** 2025-12-21
**Latest Release:** v1.7.0 (merged to main)
**Status:** Complete - all core C4 features implemented

## Goal

Add full C4 diagram support to the browser editor, using D2's native C4 features. Users should be able to create C4 diagrams with proper shapes and styling, and manipulate them interactively in the JointJS canvas.

## Quick Start (Next Session)

```bash
# View current state
git log --oneline -5
cat docs/SESSION-c4-support.md | grep -A 10 "Pending"

# Pick a feature from Pending list and create branch
git checkout -b wp-<feature-name>
```

## Background Research

### D2's C4 Support (as of v0.7.1)

D2 has first-class C4 model support including:
- `c4-person` shape - Person figure with separate head and body (better for labels)
- C4 theme (theme ID 8 "Terminal" works well for clean C4 look)
- Markdown labels for rich text descriptions
- `suspend` keyword for model reuse across views
- `d2-legend` variable for legends

**Reference:** [D2 C4 Documentation](https://d2lang.com/blog/c4/)

### Decision: Use D2's Native C4 Model

We will use D2's built-in C4 support rather than implementing our own C4 layer. This means:
1. Users write D2 syntax with `shape: c4-person` etc.
2. D2 renders to SVG with C4-specific shapes
3. Our canvas detects and renders these shapes in JointJS
4. Position/vertex metadata works the same as other shapes

### Key Architectural Decision: Extract Paths from D2's SVG

**Rather than recreating C4 shapes from scratch in JointJS, we extract the actual SVG path data from D2's output and use it directly.**

Benefits of this approach:
- **Visual fidelity**: JointJS rendering matches D2's rendering exactly
- **No guessing**: We don't need to reverse-engineer D2's shape proportions
- **Maintainability**: If D2 updates their shapes, we automatically match
- **Consistency**: Static exports and interactive canvas look identical

How it works:
1. D2 renders the diagram to SVG (the source of truth)
2. `parseD2Svg()` extracts node positions, sizes, AND path data
3. `detectShapeType()` identifies shape types from SVG structure
4. ShapeRegistry uses extracted/normalized paths for JointJS rendering
5. Paths are scaled to match the node's bounding box

## SVG Shape Analysis

### Shape Detection Patterns

From analyzing D2's SVG output, here's how to detect each shape:

| Shape | Path Count | Detection Method |
|-------|------------|------------------|
| rectangle | 0 (uses `<rect>`) | Has `<rect>` element |
| circle | 0 (uses `<ellipse>`) | Has `<ellipse>` with rx === ry |
| oval | 0 (uses `<ellipse>`) | Has `<ellipse>` with rx !== ry |
| cylinder | 2 | Second path is OPEN (no 'Z' at end) |
| **c4-person** | 2 | Both paths CLOSED, first has H/V commands |
| hexagon | 1 | lineCount === 5 (M + 5L + Z) |
| diamond | 1 | lineCount === 4 |
| person | 1 | curveCount > 3 (connected silhouette) |
| cloud | 1 | curveCount > 10 |

### C4-Person SVG Structure

```svg
<g class="[base64-encoded-id]">
  <g class="shape">
    <!-- Body: rounded rectangle -->
    <path d="M 143 54 C 143 45 152 37 160 37 H 222 C 231 37 239 46 239 54 V 93 C 239 102 230 110 222 110 H 160 C 151 110 143 101 143 93 Z"/>
    <!-- Head: circle -->
    <path d="M 191 0 C 202 0 211 9 211 21 C 211 32 202 42 191 42 C 179 42 170 32 170 21 C 170 9 179 0 191 0"/>
  </g>
  <text>Label</text>
</g>
```

Key characteristics:
- 2 `<path>` elements in `.shape` group
- First path: rounded rectangle (has H, V commands, ends with Z)
- Second path: circle (all C curves, ends closed)
- Head positioned ABOVE body (lower Y values)

### Cylinder SVG Structure (for comparison)

```svg
<g class="shape">
  <!-- Body with curved top/bottom -->
  <path d="M 270 690 C ... V 760 C ... V 690 Z"/>
  <!-- Top ellipse (OPEN path, no Z) -->
  <path d="M 270 690 C 270 714 320 714 325 714 C 331 714 380 714 380 690"/>
</g>
```

Key difference: Second path does NOT end with Z (open curve).

### Regular Person vs C4-Person

| Aspect | Regular `person` | `c4-person` |
|--------|-----------------|-------------|
| Path count | 1 | 2 |
| Structure | Connected silhouette | Separate head + body |
| Label position | Below figure | Inside body |
| Use case | Generic diagrams | C4 architecture diagrams |

## Implementation Plan

### Phase 1: Fix Shape Detection (Critical)

**Current bug:** `c4-person` has 2 paths, so it's incorrectly detected as `cylinder`.

**Fix in `detectShapeType()` function:**

```javascript
// Check paths
const paths = shapeElement.querySelectorAll('path');
if (paths.length === 2) {
    const d1 = paths[0].getAttribute('d') || '';
    const d2 = paths[1].getAttribute('d') || '';

    // Cylinder: second path is open (no Z)
    if (!d2.trim().endsWith('Z') && !d2.trim().endsWith('z')) {
        return 'cylinder';
    }

    // C4-person: both closed, first has H/V (rounded rect body)
    if ((d1.includes('H') || d1.includes('V')) &&
        (d2.trim().endsWith('Z') || d2.trim().endsWith('z'))) {
        return 'c4-person';
    }

    // Fallback for unknown 2-path shapes
    return 'cylinder';
}
```

### Phase 2: Extract and Store Path Data from D2 SVG

**Key change:** Modify `parseD2Svg()` to extract the actual SVG path data, not just positions.

```javascript
// In parseD2Svg(), when processing each node:
function extractPathData(shapeElement) {
    const paths = shapeElement.querySelectorAll('path');
    if (paths.length === 0) return null;

    // Extract all path 'd' attributes
    return Array.from(paths).map(p => p.getAttribute('d'));
}

// Add to node data:
nodes.push({
    id: decoded,
    x: bbox.x,
    y: bbox.y,
    width: bbox.width,
    height: bbox.height,
    label: extractLabel(g),
    shapeType: detectShapeType(shape),
    pathData: extractPathData(shape)  // NEW: store original paths
});
```

### Phase 3: Normalize Paths for JointJS

Create a utility to normalize D2's absolute paths to relative (0,0 origin) paths that scale with the element:

```javascript
function normalizePath(pathD, bbox) {
    // Parse the path and translate so origin is at (0,0)
    // Scale coordinates to fit within (0,0) to (1,1) for JointJS refD
    // This allows the path to scale with the element size

    // For c4-person, we need to handle both head and body paths
    // and ensure they maintain their relative positions
}
```

### Phase 4: Update ShapeRegistry to Use Extracted Paths

```javascript
ShapeRegistry.register('c4-person', (opts) => {
    // Use the path data extracted from D2's SVG
    // opts.pathData contains the original paths from D2

    if (opts.pathData && opts.pathData.length === 2) {
        // Normalize and combine the paths
        const normalizedPath = normalizePaths(opts.pathData, opts.width, opts.height);

        return new joint.shapes.standard.Path({
            position: { x: opts.x, y: opts.y },
            size: { width: opts.width, height: opts.height },
            attrs: {
                body: {
                    d: normalizedPath,
                    fill: opts.fill,
                    stroke: opts.stroke,
                    strokeWidth: 1
                },
                label: {
                    text: opts.label,
                    fill: opts.textColor,
                    fontSize: 14,
                    refX: 0.5,
                    refY: 0.7,  // Position in body area
                    textAnchor: 'middle',
                    textVerticalAnchor: 'middle'
                }
            }
        });
    }

    // Fallback to rectangle if no path data
    return ShapeRegistry.shapes['rectangle'](opts);
});
```

### Why Extract Rather Than Recreate?

| Approach | Pros | Cons |
|----------|------|------|
| **Recreate shapes** | Full control over rendering | Must reverse-engineer D2's proportions, may drift from D2's look |
| **Extract from D2** | Exact visual match, auto-updates with D2 | Slightly more complex path processing |

**We chose extraction** because visual consistency between D2's static output and our interactive canvas is critical. Users expect the diagram to look the same whether viewing a static SVG or editing in the browser.

### Phase 5: C4 Mode Flag (Already Implemented)

The `--c4` flag is already added to both `render` and `serve` commands:
- Applies Terminal theme (ID 8) by default
- Can be overridden with explicit `--theme` flag

Files already modified (uncommitted):
- `cmd/diagtool/cmd/render.go`
- `cmd/diagtool/cmd/serve.go`
- `pkg/server/server.go`
- `pkg/server/handlers.go`

### Phase 6: Sync export.html

Copy the same detection and shape registry changes to `pkg/render/export.html` for CLI export consistency.

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/server/web/dist/index.html` | Fix detectShapeType(), add c4-person to ShapeRegistry |
| `pkg/render/export.html` | Same changes as index.html |

## Test Plan

### Test Files

Example C4 diagrams already created in `examples/c4/`:
- `01-system-context.d2` - Basic context with c4-person
- `02-container.d2` - Container level with nested elements
- `03-with-layers.d2` - Multi-layer with drill-down

### Test Commands

```bash
# Build
make build

# Test CLI render with C4 mode
./bin/diagtool render examples/c4/01-system-context.d2 --c4 -o /tmp/c4-test.svg

# Test browser editor with C4 mode
./bin/diagtool serve examples/c4/01-system-context.d2 --c4

# Test shape detection
./bin/diagtool serve examples/02-shape-types.d2  # Should still work
```

### Verification Checklist

- [ ] c4-person renders as head + rounded body (not cylinder)
- [ ] Regular person still works (connected silhouette)
- [ ] Cylinder still works (database shape)
- [ ] c4-person is draggable in canvas
- [ ] c4-person position is saved/restored from metadata
- [ ] CLI export renders c4-person correctly
- [ ] --c4 flag applies Terminal theme

## Current Status

### Completed (All Released)

**v1.4.0 - C4 Support**
- [x] Research D2's C4 model support
- [x] Analyze SVG output for c4-person shape
- [x] Document detection patterns
- [x] Create example C4 diagrams
- [x] Add --c4 flag to CLI
- [x] Phase 1: Fix detectShapeType() for c4-person vs cylinder (H command check)
- [x] Phase 1b: Add c4-person shape to ShapeRegistry
- [x] Sync to export.html
- [x] Fix hexagon detection (lineCount === 5)

**v1.5.0 - Path Extraction**
- [x] Phase 2: Extract path data in parseD2Svg()
- [x] Phase 3: Create normalizePath/translatePath/scalePath utilities
- [x] Phase 4: Update ShapeRegistry to use extracted paths from D2
- [x] Exact visual fidelity between D2 output and JointJS canvas

**v1.6.0 - Color Extraction**
- [x] Extract fill color from D2 SVG
- [x] Extract stroke color from D2 SVG
- [x] Extract text color from D2 SVG
- [x] Apply colors to JointJS shapes
- [x] Stroke matches fill for seamless multi-path shapes

**v1.7.0 - Font Color & Edge Labels**
- [x] Fix CSS overriding text color (removed `fill: #000000` from `.joint-element text`)
- [x] Font colors from D2 source now apply correctly to JointJS shapes
- [x] Add edge label extraction from D2 SVG
- [x] Render edge labels on JointJS links with proper positioning
- [x] Sync all changes to export.html for CLI consistency

### Pending / Future Ideas
- [ ] C4 theme with official colors (Person=#08427b, Container=#438dd5, etc.)
- [ ] Edge label repositioning (drag to move, store in .d2meta)
- [ ] Multi-line label support (D2 markdown labels)
- [ ] Container/group support (nested D2 elements)
- [ ] Browser export buttons (PNG/PDF download from toolbar)

## Notes

- All C4 core functionality is complete and released
- Font colors from D2 source (e.g., `style.font-color: white`) now render correctly
- Edge labels are displayed at 50% position on links with white background
- Path extraction ensures visual fidelity with D2 output
