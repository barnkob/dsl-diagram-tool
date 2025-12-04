# D2 Syntax Research - WP02

Research documentation for D2 language features and syntax to inform MVP implementation.

## Overview

D2 (Declarative Diagramming) is a modern diagram scripting language that converts text to diagrams. It's designed for engineers, is open-source, and supports version control.

**Key Resources:**
- Official Documentation: https://d2lang.com/
- GitHub Repository: https://github.com/terrastruct/d2
- Interactive Playground: https://play.d2lang.com/
- Go Package: https://pkg.go.dev/oss.terrastruct.com/d2

## Core Language Features

### 1. Basic Shapes

**Default Shape (Rectangle):**
```d2
server
database
frontend
```

**Named Shapes with Labels:**
```d2
server: Web Server
database: PostgreSQL
frontend: React App
```

**Multiple Shapes:**
```d2
server; database; frontend
```

### 2. Shape Types

Specify shape using the `shape` keyword:

```d2
cloud: AWS {
  shape: cloud
}

user: Customer {
  shape: person
}

db: Database {
  shape: cylinder
}

cache: Redis {
  shape: circle
}

storage: Files {
  shape: hexagon
}
```

**Available Shapes:**
- `rectangle` (default)
- `square`
- `circle`
- `oval`
- `diamond`
- `parallelogram`
- `hexagon`
- `cylinder`
- `cloud`
- `person`
- `sql_table`
- `class`
- `code`
- `image`

### 3. Connections (Edges)

**Connection Operators:**
- `->` : Forward arrow
- `<-` : Backward arrow
- `--` : Straight line (no arrow)
- `<->` : Bidirectional arrow

**Basic Connections:**
```d2
frontend -> backend
backend -> database
api -- cache
websocket <-> server
```

**Labeled Connections:**
```d2
frontend -> backend: HTTP Request
backend -> database: SQL Query
api -- cache: Redis Protocol
websocket <-> server: WebSocket
```

**Connection Styling:**
```d2
a -> b: labeled connection {
  style: {
    stroke: red
    stroke-width: 3
    stroke-dash: 5
    animated: true
  }
}
```

### 4. Containers (Nesting)

Use curly braces to create containers:

```d2
aws: AWS Cloud {
  vpc: VPC {
    subnet1: Public Subnet {
      webserver: Web Server
    }
    subnet2: Private Subnet {
      appserver: App Server
      database: Database
    }
  }
}

# Connect across container boundaries
aws.vpc.subnet1.webserver -> aws.vpc.subnet2.appserver
```

**Infinite nesting is supported.**

### 5. Styling

Styles are applied using the `style` keyword:

**Shape Styling:**
```d2
important_node: Critical Service {
  style: {
    fill: "#ff6b6b"
    stroke: "#c92a2a"
    stroke-width: 3
    shadow: true
    border-radius: 8
    opacity: 0.9
    font-size: 14
    font-color: white
    bold: true
    3d: true
  }
}
```

**Style Properties:**
- **Visual**: `fill`, `stroke`, `stroke-width`, `stroke-dash`, `opacity`, `border-radius`
- **Effects**: `shadow`, `3d`, `multiple`, `double-border`, `animated`
- **Typography**: `font`, `font-size`, `font-color`, `bold`, `italic`, `underline`, `text-transform`

### 6. Comments

```d2
# This is a comment
server -> database # End-of-line comment

# Multi-line comments
# are done with multiple
# hash marks
```

### 7. Special Strings

Quoted strings support special characters:

```d2
"node-with-dashes"
'node with spaces'
"$$price$$" -> "???status???"
```

### 8. Dimensions

Specify explicit width and height:

```d2
logo: Company Logo {
  width: 100
  height: 50
  shape: image
}

spacer: "" {
  width: 200
  height: 20
}
```

## Special Diagram Types

### SQL Tables

Define database schemas using `sql_table` shape:

```d2
users: Users {
  shape: sql_table
  id: int {constraint: primary_key}
  email: varchar(255) {constraint: unique}
  created_at: timestamp
}

posts: Posts {
  shape: sql_table
  id: int {constraint: primary_key}
  user_id: int {constraint: foreign_key}
  title: varchar(500)
  content: text
}

users.id -> posts.user_id
```

**SQL Table Features:**
- Each key defines a row (column name)
- Primary value defines the type
- `constraint` defines SQL constraints (primary_key, foreign_key, unique, not_null)
- Connections point to specific rows with TALA/ELK layout engines

### Classes (UML)

Define class diagrams with visibility modifiers:

```d2
Shape: {
  shape: class

  # Public (+), Private (-), Protected (#)
  -x: int
  -y: int
  +area(): float
  #validate(): bool
}

Circle: {
  shape: class
  -radius: float
  +area(): float
}

Shape -> Circle: extends
```

### Sequence Diagrams

Special layout for temporal interactions:

```d2
shape: sequence_diagram

alice -> bob: Hello
bob -> alice: Hi back
alice -> charlie: How are you?
charlie -> alice: Good!
```

## Advanced Features

### Icons

Add icons from external sources:

```d2
server: Web Server {
  icon: https://icons.terrastruct.com/tech/server.svg
}
```

### Links

Make shapes clickable:

```d2
documentation: Docs {
  link: https://d2lang.com
}
```

### Markdown

Shapes can contain markdown:

```d2
readme: |md
  # Title

  - Item 1
  - Item 2

  **Bold text**
|
```

### LaTeX

Mathematical notation support:

```d2
formula: |latex
  E = mc^2
|
```

### Grid Diagrams

Organize shapes in a grid:

```d2
grid-rows: 2
grid-columns: 3

a
b
c
d
e
f
```

## Themes and Layout Engines

### Themes

D2 includes built-in themes:
- Neutral (default)
- Cool classics
- Mixed berry blue
- Grape soda
- Aubergine
- Flagship Terrastruct
- Terminal
- And more...

### Layout Engines

Three layout options:
1. **Dagre** (default, bundled) - Fast directed graph layout
2. **ELK** (bundled) - Node-link diagrams with ports
3. **TALA** (separate) - Optimized for software architecture

## Export Formats

D2 supports multiple output formats:
- SVG (vector graphics)
- PNG (raster images)
- PDF (printable documents)

## MVP Syntax Subset

For our MVP implementation, we'll focus on:

### Priority 1 (Core Features)
- [x] Basic shapes (rectangles, circles, cylinders)
- [x] Simple connections (->)
- [x] Labels on shapes and connections
- [x] Basic containers (nesting)
- [x] Comments

### Priority 2 (Essential Styling)
- [x] Shape types (shape: keyword)
- [x] Basic styling (fill, stroke, stroke-width)
- [x] Font styling (font-size, font-color, bold)

### Priority 3 (Advanced, Post-MVP)
- [ ] SQL tables
- [ ] Class diagrams
- [ ] Sequence diagrams
- [ ] Icons and images
- [ ] Markdown/LaTeX
- [ ] Grid layouts
- [ ] Advanced styling (3d, shadows, animations)

## References

- [D2 Documentation](https://d2lang.com/)
- [D2 GitHub](https://github.com/terrastruct/d2)
- [Complete Guide to D2](https://blog.logrocket.com/complete-guide-declarative-diagramming-d2/)
- [D2 Styles Reference](https://d2lang.com/tour/style/)
- [SQL Tables in D2](https://d2lang.com/tour/sql-tables/)

---

**Research completed:** 2025-12-04
**Next step:** Create example D2 files demonstrating these features
