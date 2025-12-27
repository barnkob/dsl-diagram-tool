# Tasks

## Current

## Backlog

### Testing & Quality
- [ ] Golden file tests for rendering (compare output to reference images)
- [ ] Fuzz testing for parser
- [ ] Performance benchmarks for large diagrams (100+ nodes)
- [ ] Visual regression testing setup

### Parser/Layout Enhancements
- [ ] Parser edge case handling improvements
- [ ] Enhanced error reporting
- [ ] Extended IR validation rules

### Metadata Layer
- [ ] Style overrides in `.d2meta`
- [ ] Document Git workflow with metadata files

### Visual Editor
- [ ] Property panel for editing node/edge styles
- [ ] Theme switching in UI
- [ ] Export menu in UI
- [ ] Multi-file project support

## Done

Initial migration from Work Packages - see git history for completed work:
- [x] Project setup (Go module, structure, CI/CD)
- [x] D2 syntax research and examples
- [x] Internal representation design
- [x] D2 library integration
- [x] Layout engine (Dagre)
- [x] SVG rendering
- [x] CLI tool (render, validate, version commands)
- [x] PNG/PDF export (chromedp)
- [x] Watch mode
- [x] Browser-based editor (HTTP server, WebSocket, Monaco)
- [x] JointJS interactive editor
- [x] Node dragging with position persistence
- [x] Edge vertex editing
- [x] Container/nested element support
- [x] Multi-line labels
- [x] Undo/redo
- [x] C4 diagram support
