# Session: Person Shape Proportions

**Date:** 2025-12-21
**Branch:** wp-shape-support
**Status:** Complete

## Summary

Updated person shape proportions based on user-provided SVG design (`/Users/mark/storagebox/mark/person.svg`).

## Proportions Applied

From the user's Inkscape SVG:
- **Head radius:** 13.9% of total height
- **Gap (head→body):** 3.7% of total height
- **Body width:** 100% (full width)
- **Body height:** 68.5% of total height
- **Corner radius:** 8.3% of body height

## Final Code

```javascript
// In ShapeRegistry.register('person', ...)
const headRadius = h * 0.139;
const headCX = w / 2;
const headCY = headRadius;
const gap = h * 0.037;
const bodyTop = headRadius * 2 + gap;
const bodyLeft = 0;
const bodyRight = w;
const bodyBottom = h;
const bodyHeight = bodyBottom - bodyTop;
const cornerRadius = bodyHeight * 0.083;

// Label positioned at center of body (absolute positioning)
const bodyCenterY = (bodyTop + bodyBottom) / 2;
// ...
label: { x: w / 2, y: bodyCenterY, textAnchor: 'middle', textVerticalAnchor: 'middle' }
```

## Files Updated

- `pkg/server/web/dist/index.html` - Browser editor
- `pkg/render/export.html` - CLI export template

## Notes

- Changed from `refX`/`refY` to absolute `x`/`y` for label positioning (refX/refY didn't work correctly with custom Path shapes)
- Both files kept in sync with identical person shape code
