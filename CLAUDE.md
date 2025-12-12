# CLAUDE.md

This file provides guidance to Claude Code when working on the DSL Diagram Tool.

## Project Overview

A CLI tool for rendering D2 diagrams to SVG, PNG, and PDF. Wraps the D2 library with additional features like metadata overrides and batch processing.

## Quick Reference

```bash
# Build
make build              # or: go build -o bin/diagtool ./cmd/diagtool

# Test
make test               # or: go test ./...
make verify             # fmt + vet + test

# Run
./bin/diagtool --help
./bin/diagtool render examples/basic.d2 -o output.svg
```

## Project Structure

```
cmd/diagtool/     - CLI entry point
pkg/              - Public packages
  parser/         - D2 parsing wrapper
  layout/         - Layout engine integration
  render/         - Rendering to SVG/PNG/PDF
  metadata/       - Metadata layer for overrides
internal/         - Private packages
  config/         - Configuration management
testdata/         - Test fixtures
examples/         - Example diagrams
```

## Git Workflow

This project uses work packages (WP) for development:

1. **Start work:** `git checkout -b wp##-short-name`
2. **Commit often:** Use conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`)
3. **Complete work:**
   ```bash
   git checkout main
   git merge wp##-short-name
   ```

## Key Dependencies

- **D2 Library** (`oss.terrastruct.com/d2` v0.7.1) - Core diagramming engine
- **Cobra** (`github.com/spf13/cobra`) - CLI framework

## Code Conventions

- Standard Go conventions with `gofmt`
- Table-driven tests
- Godoc comments on exported functions
- Keep functions focused and testable

## Development Notes

- See `DEVELOPMENT.md` for detailed development guide
- Project planning docs: `~/storagebox/mark/pet-projects-ideas/Projects/DSL-Diagram-Tool.md`
- Binary outputs to `bin/` (gitignored)

## Current Status

Check git branch and recent commits for current work package status:
```bash
git branch -v
git log --oneline -10
```
