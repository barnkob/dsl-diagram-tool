# D2 Examples

This directory contains example D2 diagrams demonstrating various language features.

## Example Files

### Basic Features
- **01-basic-shapes.d2** - Simple shapes and connections
- **02-shape-types.d2** - Different shape types (person, cloud, cylinder, etc.)
- **03-containers.d2** - Nested containers and cross-boundary connections

### Styling
- **04-styled-connections.d2** - Connection styling and types
- **05-styled-shapes.d2** - Shape styling with colors, borders, and fonts

### Advanced Features
- **06-sql-tables.d2** - Entity-relationship diagrams with SQL table shapes
- **07-microservices.d2** - Complex real-world microservices architecture

## Usage

These examples are intended for:
- Learning D2 syntax
- Testing parser implementation
- Validating layout algorithms
- Demonstrating tool capabilities

## Running Examples

Once the CLI is implemented (WP22), you can render these examples:

```bash
# Render to SVG
diagtool render examples/01-basic-shapes.d2 -o output.svg

# Render to PNG
diagtool render examples/02-shape-types.d2 -o output.png

# Watch mode
diagtool watch examples/ -o output/
```

## Test Fixtures

For parser testing, see `../testdata/` which contains:
- Edge cases
- Minimal valid files
- Specific feature tests
- Invalid syntax tests (for error handling)

---

**Created:** WP02 (2025-12-04)
**Updated:** WP02 (2025-12-04)
