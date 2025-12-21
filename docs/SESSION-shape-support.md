# Session Notes: Shape Support and Live Updates

**Date:** 2025-12-20
**Branch:** poc-jointjs (merged to main)
**Status:** Implementation complete

## Context

The browser editor currently renders all D2 nodes as simple rectangles using JointJS. Additionally, edits made in the D2 source editor don't reflect on the canvas properly. This session planned the implementation of proper shape support and live updates.

## Previous Work Completed This Session

Before this planning, the following was completed:
1. Merged `poc-jointjs` branch to main with JointJS migration
2. Created v1.2.0 release with browser editor features
3. Implemented client-side WYSIWYG export (SVG/PNG/PDF)
4. Added CLI export with metadata support via chromedp + JointJS
5. Fixed edge ID HTML entity encoding issue in export.html

## Requirements

### Shape Support
**Priority shapes (user selected):** Core shapes + special shapes for architectural diagrams
- rectangle (default)
- circle
- oval
- diamond
- hexagon
- person
- cylinder
- cloud

### Live Updates
**Behavior (user selected):** Re-render preserving positions
- When D2 source changes, update canvas but keep manually positioned nodes
- Smooth updates without flickering

### Extensibility
- Design must accommodate future C4 diagram support (Person, System, Container, Component)

## Technical Analysis

### Current Implementation

**File:** `pkg/server/web/dist/index.html`

The `parseD2Svg()` function (line 579) extracts nodes from D2's SVG output but only captures:
- id, x, y, width, height, label
- **Missing:** shapeType

The `createJointElements()` function hardcodes `joint.shapes.standard.Rectangle` for all nodes.

### D2 SVG Shape Detection Patterns

D2 renders different shapes using different SVG elements:

| Shape | SVG Pattern | Detection Method |
|-------|-------------|------------------|
| circle | `<ellipse rx="X" ry="Y">` | rx === ry |
| oval | `<ellipse rx="X" ry="Y">` | rx !== ry |
| rectangle | `<rect>` | Simple rect element |
| diamond | `<path>` with 4 line segments | Count L commands |
| hexagon | `<path>` with 6 line segments | Count L commands |
| cylinder | Two `<path>` elements | Count child paths |
| person | Complex `<path>` with curves | Pattern match head/shoulders |
| cloud | `<path>` with many C curves | Pattern match bumpy outline |

### JointJS Available Shapes

From `@joint/core` v4.0.4:
- `joint.shapes.standard.Rectangle`
- `joint.shapes.standard.Circle`
- `joint.shapes.standard.Ellipse`
- `joint.shapes.standard.Polygon` (for diamond, hexagon via refPoints)
- `joint.shapes.standard.Cylinder`
- `joint.shapes.standard.Path` (for custom shapes like person, cloud)

## Implementation Plan

### Step 1: Shape Detection

Add `detectShapeType(shapeElement)` function in `index.html`:

```javascript
function detectShapeType(shapeElement) {
    // Check for ellipse (circle or oval)
    const ellipse = shapeElement.querySelector('ellipse');
    if (ellipse) {
        const rx = parseFloat(ellipse.getAttribute('rx'));
        const ry = parseFloat(ellipse.getAttribute('ry'));
        return Math.abs(rx - ry) < 1 ? 'circle' : 'oval';
    }

    // Check for rect
    if (shapeElement.querySelector('rect')) {
        return 'rectangle';
    }

    // Check paths
    const paths = shapeElement.querySelectorAll('path');
    if (paths.length === 2) {
        return 'cylinder';
    }

    if (paths.length === 1) {
        const d = paths[0].getAttribute('d') || '';
        const lineCount = (d.match(/[L]/gi) || []).length;

        if (lineCount === 6) return 'hexagon';
        if (lineCount === 4) return 'diamond';

        // Complex shapes - check for curves
        const curveCount = (d.match(/[C]/gi) || []).length;
        if (curveCount > 10) return 'cloud';
        if (curveCount > 3) return 'person';
    }

    return 'rectangle';
}
```

Modify `parseD2Svg()` to include shapeType:

```javascript
nodes.push({
    id: decoded,
    x: bbox.x,
    y: bbox.y,
    width: bbox.width,
    height: bbox.height,
    label: extractLabel(g),
    shapeType: detectShapeType(shape)  // ADD THIS
});
```

### Step 2: Shape Registry

Create extensible factory pattern:

```javascript
const ShapeRegistry = {
    shapes: {},

    register(type, factory) {
        this.shapes[type] = factory;
    },

    create(type, opts) {
        const factory = this.shapes[type] || this.shapes['rectangle'];
        return factory(opts);
    }
};

// Register basic shapes
ShapeRegistry.register('rectangle', (opts) => {
    return new joint.shapes.standard.Rectangle({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: { fill: opts.fill, stroke: opts.stroke, rx: 4, ry: 4 },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

ShapeRegistry.register('circle', (opts) => {
    return new joint.shapes.standard.Circle({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: { fill: opts.fill, stroke: opts.stroke },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

ShapeRegistry.register('oval', (opts) => {
    return new joint.shapes.standard.Ellipse({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: { fill: opts.fill, stroke: opts.stroke },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

ShapeRegistry.register('diamond', (opts) => {
    return new joint.shapes.standard.Polygon({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: {
                fill: opts.fill,
                stroke: opts.stroke,
                refPoints: '0.5,0 1,0.5 0.5,1 0,0.5'
            },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

ShapeRegistry.register('hexagon', (opts) => {
    return new joint.shapes.standard.Polygon({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: {
                fill: opts.fill,
                stroke: opts.stroke,
                refPoints: '0.25,0 0.75,0 1,0.5 0.75,1 0.25,1 0,0.5'
            },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

ShapeRegistry.register('cylinder', (opts) => {
    return new joint.shapes.standard.Cylinder({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: { fill: opts.fill, stroke: opts.stroke },
            top: { fill: opts.fill, stroke: opts.stroke },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});

// Person and cloud need custom SVG paths
ShapeRegistry.register('person', (opts) => {
    return new joint.shapes.standard.Path({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: {
                fill: opts.fill,
                stroke: opts.stroke,
                d: 'M 0.5 0 C 0.7 0 0.8 0.1 0.8 0.15 C 0.8 0.2 0.7 0.25 0.5 0.25 C 0.3 0.25 0.2 0.2 0.2 0.15 C 0.2 0.1 0.3 0 0.5 0 M 0.2 0.3 L 0.8 0.3 L 0.9 0.7 L 0.7 0.7 L 0.7 1 L 0.3 1 L 0.3 0.7 L 0.1 0.7 Z'
            },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14, refY: '110%' }
        }
    });
});

ShapeRegistry.register('cloud', (opts) => {
    return new joint.shapes.standard.Path({
        position: { x: opts.x, y: opts.y },
        size: { width: opts.width, height: opts.height },
        attrs: {
            body: {
                fill: opts.fill,
                stroke: opts.stroke,
                d: 'M 0.2 0.7 C 0.05 0.7 0 0.6 0 0.5 C 0 0.35 0.1 0.3 0.2 0.3 C 0.2 0.15 0.35 0 0.5 0 C 0.65 0 0.75 0.1 0.8 0.2 C 0.9 0.2 1 0.3 1 0.45 C 1 0.6 0.9 0.7 0.8 0.7 Z'
            },
            label: { text: opts.label, fill: opts.textColor, fontSize: 14 }
        }
    });
});
```

### Step 3: Update Element Creation

In `createJointElements()`, replace hardcoded Rectangle:

```javascript
for (const node of nodes) {
    const offset = nodePositions[node.id] || { dx: 0, dy: 0 };
    const x = node.x + offset.dx;
    const y = node.y + offset.dy;

    const element = ShapeRegistry.create(node.shapeType || 'rectangle', {
        x: x,
        y: y,
        width: node.width,
        height: node.height,
        label: node.label,
        fill: canvasColors.nodeFill,
        stroke: canvasColors.nodeStroke,
        textColor: canvasColors.nodeText
    });

    element.set('nodeId', node.id);
    element.set('shapeType', node.shapeType);
    element.set('originalPosition', { x: node.x, y: node.y });

    jointElements[node.id] = element;
    graph.addCell(element);
}
```

### Step 4: Live Updates with Position Preservation

Replace full re-render with differential updates:

```javascript
function updateDiagram(newNodes, newEdges) {
    const existingNodeIds = new Set(Object.keys(jointElements));
    const newNodeIds = new Set(newNodes.map(n => n.id));

    graph.startBatch('update');

    // Remove deleted nodes
    for (const id of existingNodeIds) {
        if (!newNodeIds.has(id)) {
            jointElements[id].remove();
            delete jointElements[id];
            delete nodePositions[id];
        }
    }

    // Process nodes
    for (const node of newNodes) {
        if (jointElements[node.id]) {
            // Update existing node
            const el = jointElements[node.id];
            el.resize(node.width, node.height);
            el.attr('label/text', node.label);

            // Update original position (D2 layout changed)
            el.set('originalPosition', { x: node.x, y: node.y });

            // If shape type changed, recreate element
            if (el.get('shapeType') !== node.shapeType) {
                const currentPos = el.position();
                el.remove();

                const newEl = ShapeRegistry.create(node.shapeType, {
                    x: currentPos.x,
                    y: currentPos.y,
                    width: node.width,
                    height: node.height,
                    label: node.label,
                    fill: canvasColors.nodeFill,
                    stroke: canvasColors.nodeStroke,
                    textColor: canvasColors.nodeText
                });
                newEl.set('nodeId', node.id);
                newEl.set('shapeType', node.shapeType);
                jointElements[node.id] = newEl;
                graph.addCell(newEl);
            }
        } else {
            // Add new node
            const offset = nodePositions[node.id] || { dx: 0, dy: 0 };
            const el = ShapeRegistry.create(node.shapeType || 'rectangle', {
                x: node.x + offset.dx,
                y: node.y + offset.dy,
                width: node.width,
                height: node.height,
                label: node.label,
                fill: canvasColors.nodeFill,
                stroke: canvasColors.nodeStroke,
                textColor: canvasColors.nodeText
            });
            el.set('nodeId', node.id);
            el.set('shapeType', node.shapeType);
            el.set('originalPosition', { x: node.x, y: node.y });
            jointElements[node.id] = el;
            graph.addCell(el);
        }
    }

    // Similar logic for edges...
    // (Remove deleted, update existing, add new)

    graph.stopBatch('update');
}
```

### Step 5: Sync export.html

Copy the same shape detection and registry code to `pkg/render/export.html` for CLI export consistency.

## Files to Modify

| File | Purpose |
|------|---------|
| `pkg/server/web/dist/index.html` | Browser editor - main implementation |
| `pkg/render/export.html` | CLI export template - same shape support |

## Testing

Use `examples/02-shape-types.d2` which contains all target shapes:
- person, cloud, cylinder, circle, hexagon, diamond

Test checklist:
- [ ] All 6 shapes render correctly in browser
- [ ] Drag node, edit D2 source, verify position preserved
- [ ] Add new node in D2, verify appears without resetting others
- [ ] Delete node in D2, verify smooth removal
- [ ] Change shape type in D2, verify element updates
- [ ] CLI export renders shapes: `./bin/diagtool render examples/02-shape-types.d2 -o test.svg`
- [ ] PNG/PDF exports show shapes correctly

## Future: C4 Extension

The ShapeRegistry pattern makes C4 extension straightforward:

```javascript
ShapeRegistry.register('c4:person', {
    factory: (opts) => { /* person with C4 blue styling */ },
    defaults: { fill: '#08427B', textColor: '#FFFFFF' }
});

ShapeRegistry.register('c4:system', {
    factory: (opts) => { /* rounded rectangle with C4 styling */ },
    defaults: { fill: '#1168BD', textColor: '#FFFFFF' }
});
```

C4 metadata (technology, description) can be stored in element properties and displayed via tooltips.

## Implementation Summary

All planned features have been implemented:

### Completed

1. **Shape Detection** (`detectShapeType()` function)
   - Detects shapes from D2 SVG output by analyzing ellipse, rect, and path elements
   - Identifies: rectangle, circle, oval, diamond, hexagon, cylinder, person, cloud

2. **ShapeRegistry** (extensible factory pattern)
   - Supports all 8 target shapes with proper JointJS implementations
   - Uses `joint.shapes.standard.Rectangle`, `Circle`, `Ellipse`, `Polygon`, `Cylinder`, and `Path`
   - Custom SVG paths for person and cloud shapes

3. **parseD2Svg() Updates**
   - Now includes `shapeType` property in parsed node data

4. **createJointElements() Refactored**
   - Uses ShapeRegistry instead of hardcoded Rectangle
   - Stores shapeType on each element for change detection

5. **Live Updates** (`updateDiagram()` function)
   - Differential updates instead of full re-render
   - Preserves node positions when D2 source changes
   - Handles node add/remove/update gracefully
   - Recreates elements when shape type changes (preserving position)
   - Properly handles edges including add/remove

6. **export.html Synced**
   - Same shape detection and registry for CLI export consistency

### Testing

- Build: `make build` - success
- Verify: `make verify` - all checks pass
- CLI render: `./bin/diagtool render examples/02-shape-types.d2` - success

## Current Status (Session End)

**Date:** 2025-12-20

### What's Working
- All 8 shapes render in the browser editor
- Shape detection from D2 SVG works correctly
- Live updates preserve node positions
- Text color is black (#000000)
- CSS override fixed for text color

### Known Issues / Fine-tuning Needed

1. **Cylinder label placement** - Currently using absolute `x` and `y` coordinates (`x: opts.width / 2, y: opts.height / 2 + 10`). May still be slightly off-center. Consider trying different JointJS label attributes if still not right.

2. **Person shape proportions** - Current settings:
   - `headRadius = Math.min(w * 0.45, h * 0.25)`
   - `bodyLeft = w * 0.05, bodyRight = w * 0.95`
   - Label positioned below shape with `refX: 0.5, refY: 1, y: 8`

3. **Person/Cloud shapes** - Use dynamically generated SVG path strings based on element dimensions. If shapes look wrong at different sizes, the path generation logic in `ShapeRegistry.register('person', ...)` and `ShapeRegistry.register('cloud', ...)` may need adjustment.

### Files Modified
- `pkg/server/web/dist/index.html` - Browser editor with shape support
- `pkg/render/export.html` - CLI export with same shape support
- `docs/SESSION-shape-support.md` - This file

### To Continue
1. Test with: `./bin/diagtool serve examples/02-shape-types.d2`
2. If cylinder text is still off, try adjusting the `y` offset in the cylinder label attrs
3. If person shape needs adjustment, modify `headRadius`, `bodyLeft`, `bodyRight` values
4. Both files (index.html and export.html) must stay in sync
