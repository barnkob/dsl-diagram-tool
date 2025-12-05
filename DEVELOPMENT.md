# Development Guide

## Getting Started

### Environment Setup

1. **Install Go 1.21+**
   ```bash
   # Check your Go version
   go version
   ```

2. **Setup** (git repository already initialized)
   ```bash
   cd dsl-diagram-tool
   go mod download
   go build -o bin/diagtool ./cmd/diagtool
   ```

3. **Verify Setup**
   ```bash
   # Run tests
   go test ./...

   # Run the CLI
   ./bin/diagtool
   ```

## Project Structure

```
cmd/diagtool/       - CLI entry point (main.go)
pkg/                - Public packages (can be imported)
  parser/           - D2 parsing wrapper
  layout/           - Layout engine integration
  render/           - Rendering to SVG/PNG/PDF
  metadata/         - Metadata layer for overrides
internal/           - Private packages (project-only)
  config/           - Configuration management
testdata/           - Test fixtures
examples/           - Example diagrams
```

## Development Workflow

### Working on a Work Package

Each work package (WP) is a focused development effort. Before starting:

1. Read the work package description in `Projects/DSL-Diagram-Tool.md`
2. Check dependencies (some WPs depend on previous ones)
3. Create feature branch: `git checkout -b wp##-name`
4. Implement functionality with tests
5. Update documentation
6. Commit changes: `git commit -m "Complete WP##: description"`
7. Mark work package complete in project file

**Git Workflow:**
- Main branch: `main` (stable, completed work packages)
- Feature branches: `wp##-name` for each work package
- Commit frequently with descriptive messages
- Use conventional commit format: `feat:`, `fix:`, `docs:`, `test:`, etc.

### Running Tests

**Using Makefile (recommended):**
```bash
make test         # Run all tests
make test-v       # Run with verbose output
make test-cover   # Run with coverage report
make test-race    # Run with race detector
```

**Using Go directly:**
```bash
# All tests
go test ./...

# Specific package
go test ./pkg/parser/

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Run benchmarks (when added)
go test -bench=. ./...
```

**Coverage report:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Open in browser
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting: `go fmt ./...`
- Use `go vet` for static analysis: `go vet ./...`
- Add godoc comments for exported functions
- Keep functions focused and testable

### Building

**Using Makefile (recommended):**
```bash
make build        # Build the binary
make clean        # Remove build artifacts
make run          # Build and run
make install      # Install to GOPATH/bin
```

**Using Go directly:**
```bash
# Development build
go build -o bin/diagtool ./cmd/diagtool

# Build for specific OS/arch
GOOS=linux GOARCH=amd64 go build -o bin/diagtool-linux ./cmd/diagtool

# With version info (later)
go build -ldflags "-X main.Version=0.1.0" -o bin/diagtool ./cmd/diagtool
```

### Makefile Targets

Run `make help` to see all available targets:
```bash
make help         # Display all available commands
make build        # Build the CLI binary
make test         # Run all tests
make test-cover   # Run tests with coverage
make fmt          # Format code
make vet          # Run go vet
make lint         # Run linter (requires golangci-lint)
make verify       # Run fmt, vet, and tests
make clean        # Remove build artifacts
make all          # Clean, verify, and build
```

## Key Dependencies

### Official D2 Library (v0.7.1)

```go
import (
    "oss.terrastruct.com/d2/d2lib"
    "oss.terrastruct.com/d2/d2compiler"
    "oss.terrastruct.com/d2/d2parser"
)
```

**Key Functions:**
- `d2lib.Parse()` - Parse D2 source to AST
- `d2lib.Compile()` - Full compilation pipeline
- Multiple layout engines: Dagre (default), ELK, TALA

**Resources:**
- Go Docs: https://pkg.go.dev/oss.terrastruct.com/d2
- Examples: https://pkg.go.dev/oss.terrastruct.com/d2/docs/examples/lib

## Testing Strategy

### Unit Tests
- Test individual functions in isolation
- Use table-driven tests for multiple cases
- Mock external dependencies

### Integration Tests
- Test complete pipeline (parse → layout → render)
- Use real D2 files from testdata/
- Verify output correctness

### Golden File Tests (WP19)
- Compare rendered output to reference images
- Detect visual regressions
- Store golden files in testdata/golden/

Example test structure:
```go
func TestParser(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *IR
        wantErr bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Debugging

### Verbose Logging
```go
// Add during development
log.Printf("Debug: %+v\n", variable)
```

### Using Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug tests
dlv test ./pkg/parser

# Debug binary
dlv exec ./bin/diagtool -- render input.d2
```

## Common Tasks

### Adding a New Package

1. Create package directory: `mkdir pkg/newpackage`
2. Create package file: `pkg/newpackage/newpackage.go`
3. Add package documentation comment
4. Create test file: `pkg/newpackage/newpackage_test.go`
5. Export only necessary functions/types

### Adding Dependencies

```bash
# Add new dependency
go get <package>

# Update dependencies
go get -u ./...

# Tidy up
go mod tidy
```

### Updating D2 Library

```bash
# Check for updates
go list -m -u oss.terrastruct.com/d2

# Update to latest
go get -u oss.terrastruct.com/d2

# Update to specific version
go get oss.terrastruct.com/d2@v0.8.0
```

## Performance Considerations

- Use benchmarks to measure performance: `go test -bench=.`
- Profile CPU usage: `go test -cpuprofile=cpu.prof`
- Profile memory: `go test -memprofile=mem.prof`
- Analyze profiles: `go tool pprof cpu.prof`

## Troubleshooting

### Build Fails
```bash
# Clean build cache
go clean -cache

# Verify dependencies
go mod verify
go mod tidy
```

### Tests Fail
```bash
# Run specific test with verbose output
go test -v -run TestSpecificTest ./pkg/parser/
```

### Import Errors
- Ensure package is in go.mod
- Run `go mod tidy`
- Check import paths match module name

## Next Steps

After WP01, the next work packages are:
- **WP02**: D2 syntax research - Study D2 language features
- **WP03**: IR design - Define internal data structures
- **WP04**: D2 integration - Wrap D2 library with our abstractions

See `Projects/DSL-Diagram-Tool.md` for complete roadmap.

---

**Questions or Issues?** Document them in the project file's Notes section.
