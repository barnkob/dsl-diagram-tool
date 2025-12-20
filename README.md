# DSL Diagram Tool

A production-ready command-line tool for rendering D2 diagrams to multiple output formats (SVG, PNG, PDF). Built for developers who want version-controlled, text-based diagrams with high-quality visual output.

## Project Status

**Current Version:** v1.2.0
**Status:** ✅ Production Ready - Full CLI with SVG/PNG/PDF Export + Browser Editor

## Features

✅ **Multiple Output Formats**
- SVG (scalable vector graphics)
- PNG (high-resolution, configurable DPI)
- PDF (vector output with searchable text)

✅ **Rich Styling Options**
- 9 built-in D2 themes with dark mode variants
- Sketch/hand-drawn style
- Configurable padding and alignment
- Custom colors and shapes

✅ **Developer-Friendly**
- Watch mode for live updates during development
- Auto-format detection from file extension
- D2 syntax validation
- Comprehensive CLI with help system

✅ **Browser-Based Editor**
- Interactive diagram editing at http://localhost:8080
- Drag-and-drop node positioning
- Click edges to add/move/remove vertices (bend points)
- Real-time sync with D2 source file
- Layout changes persist to `.d2meta` files

✅ **High Quality Output**
- Proper font rendering (via headless Chrome)
- High-DPI PNG support (default 3x, configurable)
- Vector PDF with embedded fonts
- Preserves D2 styling and themes

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
- **CLI Framework:** cobra (v1.10.2)
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

## Installation

### Prerequisites

- **Go 1.21 or later** - [Download Go](https://go.dev/dl/)
- **Chrome or Chromium** - Required for PNG/PDF export (uses headless browser)
- **Git** - For cloning the repository

### Build from Source

```bash
# Clone the repository
git clone https://github.com/mark/dsl-diagram-tool.git
cd dsl-diagram-tool

# Install dependencies
go mod download

# Build the CLI
make build
# OR: go build -o bin/diagtool ./cmd/diagtool

# Verify installation
./bin/diagtool version

# Run tests
make test
# OR: go test ./...
```

### Add to PATH (Optional)

```bash
# Move binary to a location in your PATH
sudo mv bin/diagtool /usr/local/bin/

# Now you can use it from anywhere
diagtool --help
```

## Quick Start

### Basic Usage

```bash
# Render to SVG (default)
diagtool render diagram.d2

# Render to PNG (auto-detected from extension)
diagtool render diagram.d2 -o diagram.png

# Render to PDF
diagtool render diagram.d2 -o diagram.pdf

# Validate D2 syntax
diagtool validate diagram.d2
```

### Common Examples

```bash
# High-resolution PNG (4x DPI)
diagtool render diagram.d2 -o diagram.png --pixel-density 4

# Dark mode with sketch style
diagtool render diagram.d2 -o output.svg --dark --sketch

# PDF with custom theme
diagtool render diagram.d2 -o output.pdf --theme 5

# Watch mode for live updates
diagtool render diagram.d2 --watch -o output.svg

# Custom padding and no centering
diagtool render diagram.d2 --padding 200 --no-center
```

### All Available Options

```bash
diagtool render [input.d2] [flags]

Flags:
  -o, --output string         Output file path (auto-detects format from extension)
  -f, --format string         Output format: svg, png, pdf (default "svg")
  -t, --theme int             Theme ID 0-8 (default 0)
  -d, --dark                  Use dark mode theme
  -s, --sketch                Use sketch/hand-drawn style
  -p, --padding int           Padding around diagram in pixels (default 100)
      --no-center             Don't center the diagram
      --pixel-density int     PNG pixel density/DPI multiplier (default 3)
  -w, --watch                 Watch mode: auto-regenerate on file changes
  -h, --help                  Help for render command
```

### Output Format Details

**SVG** - Scalable vector graphics, perfect for web and presentations
- Smallest file size
- Infinitely scalable
- Native D2 output

**PNG** - High-resolution raster images
- Default 3x pixel density for crisp output
- Configurable DPI (1x standard, 2x retina, 3-4x high-DPI)
- Uses headless Chrome for proper font rendering

**PDF** - Print-ready documents with vector graphics
- Searchable text (fonts embedded)
- Vector quality (scales perfectly)
- A4 paper size with 0.4" margins
- Uses headless Chrome for consistent rendering

## Examples

### Create a Simple Diagram

Create a file `architecture.d2`:
```d2
users: Users {
  shape: person
}

api: API Server {
  shape: rectangle
}

database: PostgreSQL {
  shape: cylinder
}

cache: Redis {
  shape: circle
}

users -> api: HTTP requests
api -> database: SQL queries
api -> cache: Get/Set
```

Render it:
```bash
# SVG for web
diagtool render architecture.d2 -o architecture.svg

# High-res PNG for documentation
diagtool render architecture.d2 -o architecture.png --pixel-density 4

# PDF for printing
diagtool render architecture.d2 -o architecture.pdf --dark
```

### Watch Mode During Development

```bash
# Start watch mode
diagtool render architecture.d2 --watch -o output.svg

# Edit architecture.d2 in your editor
# Output automatically updates on save
# Press Ctrl+C to stop watching
```

### Browser-Based Editor

The `serve` command launches an interactive browser-based editor powered by [JointJS](https://www.jointjs.com/):

```bash
# Start the editor server
diagtool serve diagram.d2

# Opens http://localhost:8080 in your browser
```

**Interactive Features:**
- **Drag nodes** - Click and drag any node to reposition it
- **Add vertices** - Click an edge to select it, then click on the edge path to add a bend point
- **Move vertices** - Drag the vertex circles to reshape edge routing
- **Remove vertices** - Double-click a vertex to remove it
- **Real-time sync** - Changes are saved automatically to a `.d2meta` file
- **Export** - Click SVG/PNG/PDF buttons or use keyboard shortcuts

**Keyboard Shortcuts** (when focus is on the canvas):
| Key | Action |
|-----|--------|
| `1` | Export as SVG |
| `2` | Export as PNG |
| `3` | Export as PDF |
| `R` | Reset layout |
| `Escape` | Deselect edges |
| `Ctrl+S` | Save file (in editor) |

The D2 source file remains unchanged - all layout customizations are stored separately in `.d2meta` files.

### All Commands

```bash
# Render command
diagtool render <input.d2> [flags]

# Serve command (browser editor)
diagtool serve <input.d2> [--port 8080]

# Validate command
diagtool validate <input.d2> [-v|--verbose]

# Version information
diagtool version

# Help
diagtool --help
diagtool render --help
```

## Development

### Building and Testing

```bash
# Using Makefile (recommended)
make build        # Build the binary
make test         # Run all tests
make test-cover   # Run tests with coverage
make verify       # Run fmt, vet, and tests
make clean        # Remove build artifacts
make help         # See all available commands

# Or using Go directly
go build -o bin/diagtool ./cmd/diagtool
go test ./...
go test -cover ./...
```

### Project Structure

```
dsl-diagram-tool/
├── cmd/diagtool/          # CLI entry point
│   └── cmd/               # Cobra commands
├── pkg/
│   ├── parser/            # D2 parsing (wraps official D2 lib)
│   ├── layout/            # Layout algorithms (Dagre)
│   ├── render/            # Rendering to SVG/PNG/PDF
│   ├── ir/                # Internal representation
│   └── metadata/          # Position/style override layer
├── examples/              # Example D2 diagrams
├── testdata/              # Test fixtures
└── .github/workflows/     # CI/CD automation
```

### Test Coverage

- **CLI Commands**: 48 tests, 60.7% coverage
- **Render Package**: Comprehensive unit and integration tests
- **Parser Package**: D2 integration tests
- **Overall**: All tests passing ✅

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development guide.

## Technical Details

### Architecture

```
D2 File → D2 Parser → Internal Representation → Layout Engine → Renderer → Output
                                                       ↓
                                                Metadata Layer
                                        (position & style overrides)
```

### Key Technologies

- **Language**: Go 1.21+
- **D2 Library**: oss.terrastruct.com/d2 v0.7.1 (official D2 library)
- **CLI Framework**: Cobra v1.10.2
- **Layout Engine**: D2's Dagre implementation
- **PNG/PDF Rendering**: chromedp (headless Chrome)

### Why These Choices?

- **Official D2 Library**: Mature, actively maintained (22.6k+ stars), handles complex D2 syntax
- **Go**: Fast execution, single binary distribution, cross-platform
- **Headless Chrome**: Proper font rendering, consistent output across platforms
- **Metadata Layer**: Future support for visual editing while maintaining text-based source

## Troubleshooting

### Chrome/Chromium Not Found

PNG and PDF export require Chrome or Chromium:

**macOS:**
```bash
brew install --cask google-chrome
# OR
brew install chromium
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# Fedora
sudo dnf install chromium
```

**Windows:**
Download from [google.com/chrome](https://www.google.com/chrome/)

### Build Errors

```bash
# Clean build cache
go clean -cache

# Verify dependencies
go mod verify
go mod tidy

# Rebuild
make clean && make build
```

## Roadmap

### v1.0 ✅
- SVG, PNG, PDF export
- Watch mode
- Comprehensive CLI
- Full test coverage

### v1.1 ✅
- Browser-based editor with live preview
- Server mode with WebSocket sync
- Metadata layer for position overrides

### v1.2 (Current) ✅
- JointJS-based interactive editor
- Drag-and-drop node positioning
- Edge vertex manipulation (bend points)
- Real-time layout persistence

### Future Enhancements
- Export buttons in browser editor
- Keyboard shortcuts for editor
- Orthogonal edge routing
- Additional layout engines (ELK, TALA)
- Batch processing
- Custom paper sizes for PDF

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright (c) 2025 Mark Barnkob

## Acknowledgments

- [D2](https://d2lang.com/) - The excellent diagramming language by Terrastruct
- [JointJS](https://www.jointjs.com/) - JavaScript diagramming library for browser editor
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [chromedp](https://github.com/chromedp/chromedp) - Headless Chrome automation

## Resources

- **D2 Language:** https://d2lang.com/
- **D2 GitHub:** https://github.com/terrastruct/d2
- **D2 Go Docs:** https://pkg.go.dev/oss.terrastruct.com/d2
- **Project Origin:** See Ideas vault in parent directory

---

**Version:** 1.2.0
**Last Updated:** 2025-12-20
**Status:** Production Ready
