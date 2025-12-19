# Session Notes: WP34 Drag-Drop & Vertices

**Date:** 2025-12-16 (updated 2025-12-19)
**Branch:** `wp34-drag-drop` → `poc-jointjs`
**Status:** Completed - Migrated to JointJS

## What Was Accomplished

### 1. Waypoint Bug Fixes

Fixed critical coordinate transformation bugs in `pkg/server/web/dist/index.html`:

- **Root cause:** Double-compensation for pan/zoom. Code used `getScreenCTM().inverse()` (which already converts to SVG coordinates) then incorrectly applied additional `- panX / zoom` compensation.

- **Fixed locations:**
  - Line ~1156: Click handler for adding waypoints
  - Line ~1178: 'V' key handler for adding waypoints
  - Line ~1532: Drag handler for moving waypoints

### 2. UX Improvements

- Changed waypoint creation from **single-click to double-click**
- Implemented **inactive waypoint state**:
  - New waypoints appear as dashed circles
  - Edge path unchanged until user drags the waypoint
  - On drag start, waypoint activates and edge routes through it
  - Only active waypoints are saved to server

### 3. Documentation Created

- `CLAUDE.md` - Added Browser Editor Features section with waypoints overview
- `docs/DESIGN-waypoints.md` - Detailed implementation documentation
- `docs/DESIGN-layout-metadata.md` - Structurizr research and comparison

## Current State

### Files Modified (Uncommitted)
- `pkg/server/handlers.go`
- `pkg/server/metadata.go`
- `pkg/server/server.go`
- `pkg/server/web/dist/index.html`

### New Files
- `docs/DESIGN-waypoints.md`
- `docs/DESIGN-layout-metadata.md`
- `docs/SESSION-wp34-waypoints.md` (this file)
- `examples/03-containers.d2meta`
- `examples/test-nested.d2`
- `examples/test-nested.d2meta`

## Next Steps

### Step 1: Test Current Changes
```bash
./bin/diagtool edit examples/01-basic-shapes.d2
```
- Double-click edge → inactive waypoint appears
- Drag waypoint → activates, edge bends
- Verify coordinates are correct
- Test delete (X button) and routing toggle (R key)

### Step 2: Commit & Merge Waypoint Fixes
```bash
git add -A
git commit -m "fix: Waypoint coordinate bugs and UX improvements (WP34)

- Fix double pan/zoom compensation in coordinate transformations
- Change waypoint creation from click to double-click
- Add inactive waypoint state (edge unchanged until first drag)
- Add CSS for inactive waypoint visual style
- Update design documentation"

git checkout main
git merge wp34-drag-drop
```

### Step 3: Implement Smart Merging (New Branch)
```bash
git checkout -b wp35-smart-merge
```

**Goal:** Replace source hash clearing with element-based matching

**Implementation plan:**
1. Parse D2 source to extract node/edge IDs
2. On source change, compare old vs new IDs
3. Preserve positions for nodes that still exist
4. Preserve waypoints for edges where both endpoints exist
5. Only clear layout for removed elements

**Key files to modify:**
- `pkg/server/metadata.go` - Add smart merge logic
- `pkg/server/server.go` - Call smart merge on file change

### Step 4: JointJS PoC (Separate Branch)
```bash
git checkout main
git checkout -b poc-jointjs
```

**Goal:** Evaluate JointJS integration feasibility

**PoC scope:**
1. Parse D2 SVG output into JointJS paper
2. Enable drag-drop on nodes
3. Sync position changes back to metadata
4. Re-render D2 and re-sync to JointJS
5. Evaluate complexity vs benefit

**Key questions to answer:**
- How well does JointJS handle arbitrary SVG shapes?
- What's the performance overhead of dual-model sync?
- Is the UX improvement worth the complexity?

## Architecture Notes

### Current Metadata Flow
```
D2 Source → D2 Render → SVG → DOM manipulation → User interaction
                                                        ↓
                                               Save to .d2meta
                                                        ↓
                                               Persist positions/waypoints
```

### Structurizr Comparison
| Feature | Structurizr | Current |
|---------|-------------|---------|
| Positions | Absolute (x, y) | Offset (dx, dy) |
| Change handling | Smart merge | Source hash (clear all) |
| Waypoints | vertices array | waypoints array ✓ |
| Routing | Direct, Curved, Orthogonal | Direct, Orthogonal |

### Recommended Priority
1. **Smart merging** - High value, medium effort, low risk
2. **Undo/redo** - Medium value, medium effort, low risk
3. **JointJS** - Unknown value, high effort, medium risk (needs PoC)

## Commands Reference

```bash
# Build
make build

# Test
make test

# Run editor
./bin/diagtool edit examples/01-basic-shapes.d2

# Check git status
git status
git diff --stat
```

## Related Documents

- [CLAUDE.md](../CLAUDE.md) - Project overview and waypoints summary
- [DESIGN-waypoints.md](./DESIGN-waypoints.md) - Waypoints implementation details
- [DESIGN-layout-metadata.md](./DESIGN-layout-metadata.md) - Structurizr research
