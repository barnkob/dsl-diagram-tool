# DSL Diagram Tool

A hybrid diagramming tool that bridges text-based diagram creation (D2 DSL) with visual editing capabilities. Enables version-controlled diagram-as-code while preserving human layout decisions through a metadata layer.

## Project Status

**Current Version:** 0.1.0-dev (WP08 completed)
**Status:** 🏗️ Layout Engine Implementation

## Overview

This tool enables software architects and developers to:
- Write diagrams in D2 DSL (text-based, version controlled)
- Render diagrams to multiple formats (SVG, PNG, PDF)
- Apply custom position and style overrides via metadata
- Eventually edit diagrams visually while maintaining DSL source

## Architecture

```
D2 File → Parser (D2 lib) → Internal Representation → Layout Engine → Renderer → Output
                                                            ↓
                                                     Metadata Layer
                                            (position & style overrides)
```

### Tech Stack

- **Language:** Go 1.21+
- **D2 Library:** oss.terrastruct.com/d2 (v0.7.1)
- **CLI Framework:** To be added (cobra - WP20)
- **Layout Engines:** D2's Dagre, ELK, TALA
- **Output Formats:** SVG, PNG, PDF

## Project Structure

```
dsl-diagram-tool/
├── cmd/diagtool/          # CLI entry point
├── pkg/
│   ├── parser/            # D2 parsing (wraps official D2 lib)
│   ├── layout/            # Layout algorithms
│   ├── render/            # Rendering to various formats
│   └── metadata/          # Position/style override layer
├── internal/config/       # Internal configuration
├── testdata/              # Test fixtures and sample diagrams
├── examples/              # Example diagrams and usage
└── .github/workflows/     # CI/CD automation
```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Installation

```bash
# Navigate to the project (already initialized as git repo)
cd dsl-diagram-tool

# Install dependencies
go mod download

# Build the CLI
go build -o bin/diagtool ./cmd/diagtool

# Run tests
go test ./...
```

**Note:** This project is a git repository. The `.git` folder is located at the project root. GitHub Actions workflows will be functional once you push to a remote repository (e.g., GitHub).

### Building

```bash
# Using Makefile (recommended)
make build        # Build the binary
make test         # Run tests
make test-cover   # Run tests with coverage
make verify       # Run fmt, vet, and tests
make help         # See all available commands

# Or using Go directly
go build -o bin/diagtool ./cmd/diagtool
go test ./...
```

## Work Packages

Development is organized into 31 incremental work packages across 5 phases:

### Phase 1: Foundation & Core Parsing (WP01-07)
- [x] **WP01**: Project setup, repository, dependencies ✅
- [x] **WP02**: D2 syntax research and example collection ✅
- [x] **WP03**: Internal Representation design ✅
- [x] **WP04**: D2 integration and D2→IR converter ✅
- [ ] **WP05**: Parser refinements (if needed)
- [ ] **WP06**: IR validation enhancements (if needed)
- [ ] **WP07**: Parser test suite expansion

### Phase 2: Layout Engine (WP08-13)
- [x] **WP08**: Layout engine integration (Dagre) ✅
### Phase 3: Rendering Engine (WP14-19)
### Phase 4: CLI Tool (WP20-26)
### Phase 5: Metadata Layer (WP27-31)

See [project documentation](../Projects/DSL-Diagram-Tool.md) for complete work package breakdown.

## Key Decisions

- **Using Official D2 Library:** Instead of building a custom parser, we leverage terrastruct's mature D2 library (22.6k stars, actively maintained)
- **Go Language:** Fast execution, single binary distribution, excellent CLI tooling
- **Metadata Separation:** Position/style overrides stored separately from DSL source for clean Git workflows

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/parser/
```

## Contributing

This is currently a personal project in early development. Work packages are being implemented sequentially.

## License

To be determined

## Resources

- **D2 Language:** https://d2lang.com/
- **D2 GitHub:** https://github.com/terrastruct/d2
- **D2 Go Docs:** https://pkg.go.dev/oss.terrastruct.com/d2
- **Project Origin:** See Ideas vault in parent directory

---

**Last Updated:** 2025-12-09 (WP04 completed)
