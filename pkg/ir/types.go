package ir

// ShapeType represents the type of a node shape.
type ShapeType string

// Shape types supported by the IR.
const (
	// Basic shapes
	ShapeRectangle     ShapeType = "rectangle"
	ShapeSquare        ShapeType = "square"
	ShapeCircle        ShapeType = "circle"
	ShapeOval          ShapeType = "oval"
	ShapeDiamond       ShapeType = "diamond"
	ShapeParallelogram ShapeType = "parallelogram"
	ShapeHexagon       ShapeType = "hexagon"

	// Special shapes
	ShapePerson   ShapeType = "person"
	ShapeCloud    ShapeType = "cloud"
	ShapeCylinder ShapeType = "cylinder"

	// Container
	ShapeContainer ShapeType = "container"

	// Advanced shapes (post-MVP)
	ShapeSQLTable ShapeType = "sql_table"
	ShapeClass    ShapeType = "class"
	ShapeCode     ShapeType = "code"
	ShapeImage    ShapeType = "image"
)

// Direction represents the direction of an edge.
type Direction string

// Edge directions.
const (
	DirectionForward  Direction = "forward"  // ->
	DirectionBackward Direction = "backward" // <-
	DirectionBoth     Direction = "both"     // <->
	DirectionNone     Direction = "none"     // --
)

// PositionSource indicates how a position was determined.
type PositionSource string

// Position sources.
const (
	PositionSourceLayoutEngine PositionSource = "layout_engine"
	PositionSourceMetadata     PositionSource = "metadata"
	PositionSourceManual       PositionSource = "manual"
)
