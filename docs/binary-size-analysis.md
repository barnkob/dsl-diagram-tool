# Binary Size Analysis

**Date:** 2025-12-12
**Current Binary Size:** 31MB (unstripped) → 25MB (stripped)

## Root Cause

The large binary size is primarily due to the D2 library and its extensive dependency tree:

### D2 Library Dependencies (84 total)
The D2 library (`oss.terrastruct.com/d2`) includes:

- **JavaScript Engine** (`github.com/dop251/goja`) - For extensibility and scripting
- **Browser Automation** (`github.com/playwright-community/playwright-go`) - For advanced rendering/screenshots
- **Image Processing** - Multiple image manipulation libraries
- **PDF Generation** (`github.com/jung-kurt/gofpdf`)
- **Syntax Highlighting** (`github.com/alecthomas/chroma/v2`)
- **Font Rendering** (`github.com/golang/freetype`)
- **WebSocket Support** (`github.com/coder/websocket`)

### What We Actually Use
Our CLI only uses a subset of D2's functionality:
- `d2compiler` - Parsing D2 files
- `d2graph` - Internal graph representation
- `d2layouts/d2dagrelayout` - Dagre layout engine
- `d2renderers/d2svg` - SVG rendering
- `lib/textmeasure` - Text measurement

However, Go's linker includes all transitive dependencies from these packages.

## Optimization Applied

### 1. Build Flags (6MB reduction)
```bash
go build -ldflags="-s -w" -trimpath -o bin/diagtool ./cmd/diagtool
```

**Flags:**
- `-s` - Strip symbol table
- `-w` - Strip DWARF debugging info
- `-trimpath` - Remove file system paths from binary

**Result:** 31MB → 25MB (19% reduction)

### 2. Comparison with Official D2 Binary
The official D2 CLI is also large (~30MB) for the same reasons. This is the trade-off for having a feature-rich library.

## Further Optimization Options

### Option A: UPX Compression (Not Recommended for Production)
UPX can compress the binary to ~8-10MB but:
- Slower startup time (decompression overhead)
- May trigger false positives in antivirus software
- Not suitable for production distribution

```bash
# Install UPX (macOS)
brew install upx

# Compress binary
upx --best --lzma bin/diagtool
```

### Option B: Plugin Architecture (Future Work)
Move layout engines and renderers to separate plugins loaded at runtime. Only the core parser would be in the main binary.

**Pros:**
- Smaller core binary (~5-10MB)
- Users only download needed plugins

**Cons:**
- More complex deployment
- Runtime plugin loading overhead
- Significant refactoring required

### Option C: Custom D2 Parser (Not Recommended)
Write our own D2 parser without the full D2 library.

**Pros:**
- Full control over dependencies
- Potentially smaller binary

**Cons:**
- Massive implementation effort
- Maintenance burden
- Feature parity difficult
- Defeats purpose of using established library

### Option D: Accept Current Size (Recommended)
25MB is reasonable for a modern CLI tool with full rendering capabilities:
- Docker CLI: 45MB
- Terraform: 100MB+
- kubectl: 50MB
- Official D2: 30MB

## Recommendations

1. **Update Makefile** - Use optimized build flags by default
2. **Document size** - Add note to README about binary size and reasoning
3. **Monitor dependencies** - Periodically check if D2 reduces their footprint
4. **Consider Option B** - If binary size becomes a critical issue (enterprise firewalls, etc.)

## Current Build Settings

```makefile
# Optimized production build
build:
	go build -ldflags="-s -w" -trimpath -o bin/diagtool ./cmd/diagtool

# Development build with debug symbols
build-dev:
	go build -o bin/diagtool ./cmd/diagtool
```
