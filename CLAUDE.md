# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI tool for rendering D2 diagrams to SVG, PNG, and PDF, plus a browser-based interactive editor. Wraps the D2 library with metadata overlays for position/vertex customization.

## Quick Reference

```bash
# Build
make build              # Optimized build to bin/diagtool
make build-dev          # Debug build with symbols

# Test
make test               # Run all tests
make verify             # fmt + vet + test
go test ./pkg/parser/   # Test specific package
go test -v -run TestName ./pkg/render/  # Run single test

# Run
./bin/diagtool render examples/01-basic-shapes.d2 -o output.svg
./bin/diagtool serve examples/01-basic-shapes.d2   # Browser editor at :8080
```

## Architecture

```
D2 Source (.d2) → D2 Library → SVG → Browser Editor (JointJS) → User Interactions
                                ↓                                      ↓
                         Metadata (.d2meta)  ←←←←←←←←←←←←←←←←←←  WebSocket sync
```

**Key architectural decisions:**
- D2 source files are read-only; layout customizations go to `.d2meta` sidecar files
- Browser editor parses D2-rendered SVG into JointJS for interactive manipulation
- Positions stored as offsets (dx, dy) from auto-layout, not absolute coordinates
- Source hash in `.d2meta` triggers full layout clear when D2 content changes

## Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/server/` | HTTP server, WebSocket handlers, metadata persistence |
| `pkg/render/` | D2→SVG/PNG/PDF rendering, `jointjs.go` for shape parsing |
| `pkg/parser/` | D2 source parsing wrapper |
| `pkg/ir/` | Internal representation types (Node, Edge, Diagram) |

### Server Package (`pkg/server/`)

- `server.go` - HTTP server setup, file watching
- `handlers.go` - REST API and WebSocket message handlers
- `metadata.go` - `Metadata` struct with `Positions`, `Vertices` maps; `.d2meta` file I/O
- `web/dist/index.html` - Single-file JointJS frontend

**WebSocket message types:** `render`, `positions`, `vertices`, `file-changed`, `reset-layout`

## Git Workflow

Branch naming:
- `feat/short-description` for features
- `fix/issue-name` for bug fixes

Conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`

## Workflow

- Check `docs/TASKS.md` for current work
- Update TASKS.md when completing items
- Keep commits focused and atomic
- Ask before making significant architectural changes
- See `docs/DECISIONS.md` for architectural context

## Browser Editor (JointJS)

See `docs/DESIGN-vertices.md` for detailed design.

**Data flow:**
1. Server renders D2→SVG and sends to browser
2. Frontend parses SVG to extract node positions/sizes, edge connections, colors, and labels
3. JointJS graph built from parsed data + stored metadata
4. User drags nodes/vertices → changes sent via WebSocket → saved to `.d2meta`

**Interaction model:**
- Drag nodes to reposition
- Click edge to select, click edge path to add vertex, drag to move, double-click to remove

**Supported shapes:** rectangle, circle, oval, diamond, hexagon, cylinder, person, c4-person, cloud

**Style extraction from D2:**
- `style.fill` → shape fill color
- `style.stroke` → shape stroke color
- `style.font-color` → text/label color
- Edge labels from connection definitions (e.g., `a -> b: label text`)

## C4 Diagram Support

Use the `--c4` flag for C4 architecture diagrams:
```bash
./bin/diagtool serve examples/c4/04-with-theme.d2 --c4
./bin/diagtool render examples/c4/04-with-theme.d2 --c4 -o output.svg
```

The `--c4` flag:
- Applies the Terminal theme (ID 8) for a clean professional look
- Injects C4 style classes with Structurizr's conventional color scheme

**C4 Theme Classes** (available when using `--c4`):

| Class | Purpose | Color |
|-------|---------|-------|
| `c4-person` | Person actor | Dark blue (#08427b) |
| `c4-system` | Software System | Medium blue (#1168bd) |
| `c4-container` | Container | Light blue (#438dd5) |
| `c4-component` | Component | Lightest blue (#85bbf0) |
| `c4-external` | External System | Gray (#999999) |
| `c4-external-person` | External Person | Gray (#999999) |

**Usage in D2:**
```d2
customer: Customer {
  class: c4-person
}
banking: Internet Banking System {
  class: c4-system
}
mainframe: Legacy System {
  class: c4-external
}
```

D2's native `c4-person` shape renders correctly with separate head and body.

## Dependencies

- `oss.terrastruct.com/d2` v0.7.1 - D2 diagramming library
- `github.com/spf13/cobra` - CLI framework
- `github.com/chromedp/chromedp` - Headless Chrome for PNG/PDF export

## Useful Commands

```bash
# Check current work package status
git branch -v
git log --oneline -10

# Run browser editor with example
make build && ./bin/diagtool serve examples/01-basic-shapes.d2
```
