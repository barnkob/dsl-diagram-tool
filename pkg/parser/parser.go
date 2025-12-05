package parser

import "github.com/mark/dsl-diagram-tool/pkg/ir"

// Package parser provides D2 diagram parsing capabilities.
// This package wraps the official terrastruct/d2 library.
// Implementation will be added in WP04-06.

// Parser is the interface for diagram parsers.
// Different DSL parsers (D2, PlantUML, Mermaid) implement this interface.
type Parser interface {
	// Parse converts DSL source code to internal representation.
	Parse(source string) (*ir.Diagram, error)
}

// D2Parser wraps the official terrastruct/d2 library.
// Implementation in WP04-06.
type D2Parser struct {
	// Configuration will be added
}
