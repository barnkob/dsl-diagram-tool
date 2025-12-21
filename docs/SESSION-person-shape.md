# Session: Person Shape Proportions

**Date:** 2025-12-21
**Branch:** wp-shape-support
**Status:** Waiting for user SVG input

## Context

Working on improving the person shape in the JointJS browser editor. The shape consists of:
- A circle for the head
- A rounded rectangle for the body
- Label text centered in the body

## Current State

The person shape proportions have been adjusted but need further refinement. Current settings:
- Head: 25% of height (radius = `h * 0.125`)
- Body: 60% of width, centered (from 20% to 80%)
- Label: Centered vertically in the body rectangle

## Next Step

User will create an SVG file with the desired person shape proportions:
- A circle for the head
- A rectangle for the body

Once the SVG is provided, extract the coordinates and calculate:
1. Head radius as percentage of total height
2. Body width as percentage of total width
3. Body top position relative to head
4. Any gap/overlap between head and body

## Files to Update

Both files must stay in sync:
- `pkg/server/web/dist/index.html` - Browser editor (lines ~707-756)
- `pkg/render/export.html` - CLI export template (lines ~232-281)

## Current Person Shape Code

```javascript
// In ShapeRegistry.register('person', ...)
const headRadius = Math.min(w * 0.30, h * 0.125);
const headCX = w / 2;
const headCY = headRadius;
const bodyTop = headRadius * 2 - 2; // Overlap slightly with head
const bodyLeft = w * 0.20;
const bodyRight = w * 0.80;
const bodyBottom = h;
const cornerRadius = 6;

// Label positioned at vertical center of body
const bodyCenterY = (bodyTop + bodyBottom) / 2;
const labelRefY = bodyCenterY / h;
```

## To Continue

1. User provides SVG file path with person shape design
2. Read SVG and extract circle (head) and rect (body) coordinates
3. Calculate proportions relative to overall bounding box
4. Update both index.html and export.html
5. Test with: `./bin/diagtool serve examples/02-shape-types.d2`
