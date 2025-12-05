package ir

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks the diagram for structural and semantic errors.
func (d *Diagram) Validate() []error {
	var errors []error

	// Check for duplicate node IDs
	nodeIDs := make(map[string]bool)
	for _, node := range d.Nodes {
		if node.ID == "" {
			errors = append(errors, ValidationError{
				Field:   "node.ID",
				Message: "node ID cannot be empty",
			})
			continue
		}

		if nodeIDs[node.ID] {
			errors = append(errors, ValidationError{
				Field:   "node.ID",
				Message: fmt.Sprintf("duplicate node ID: %s", node.ID),
			})
		}
		nodeIDs[node.ID] = true
	}

	// Check for duplicate edge IDs
	edgeIDs := make(map[string]bool)
	for _, edge := range d.Edges {
		if edge.ID == "" {
			errors = append(errors, ValidationError{
				Field:   "edge.ID",
				Message: "edge ID cannot be empty",
			})
			continue
		}

		if edgeIDs[edge.ID] {
			errors = append(errors, ValidationError{
				Field:   "edge.ID",
				Message: fmt.Sprintf("duplicate edge ID: %s", edge.ID),
			})
		}
		edgeIDs[edge.ID] = true
	}

	// Validate edges reference existing nodes
	for _, edge := range d.Edges {
		if edge.Source == "" {
			errors = append(errors, ValidationError{
				Field:   "edge.Source",
				Message: fmt.Sprintf("edge %s has empty source", edge.ID),
			})
		} else if !nodeIDs[edge.Source] {
			errors = append(errors, ValidationError{
				Field:   "edge.Source",
				Message: fmt.Sprintf("edge %s references non-existent source node: %s", edge.ID, edge.Source),
			})
		}

		if edge.Target == "" {
			errors = append(errors, ValidationError{
				Field:   "edge.Target",
				Message: fmt.Sprintf("edge %s has empty target", edge.ID),
			})
		} else if !nodeIDs[edge.Target] {
			errors = append(errors, ValidationError{
				Field:   "edge.Target",
				Message: fmt.Sprintf("edge %s references non-existent target node: %s", edge.ID, edge.Target),
			})
		}
	}

	// Validate container references
	for _, node := range d.Nodes {
		if node.Container != "" && !nodeIDs[node.Container] {
			errors = append(errors, ValidationError{
				Field:   "node.Container",
				Message: fmt.Sprintf("node %s references non-existent container: %s", node.ID, node.Container),
			})
		}

		// Validate hierarchical ID matches container
		if node.Container != "" {
			expectedPrefix := node.Container + "."
			if !strings.HasPrefix(node.ID, expectedPrefix) {
				errors = append(errors, ValidationError{
					Field:   "node.ID",
					Message: fmt.Sprintf("node %s has container %s but ID doesn't match hierarchy", node.ID, node.Container),
				})
			}
		}
	}

	// Validate style values
	for _, node := range d.Nodes {
		errors = append(errors, validateStyle(node.Style, fmt.Sprintf("node %s", node.ID))...)
	}
	for _, edge := range d.Edges {
		errors = append(errors, validateStyle(edge.Style, fmt.Sprintf("edge %s", edge.ID))...)
	}

	return errors
}

// validateStyle checks style values are within valid ranges.
func validateStyle(style Style, context string) []error {
	var errors []error

	// Validate opacity
	if style.Opacity < 0.0 || style.Opacity > 1.0 {
		if style.Opacity != 0 { // 0 means unset
			errors = append(errors, ValidationError{
				Field:   context + ".style.Opacity",
				Message: fmt.Sprintf("opacity must be between 0.0 and 1.0, got %f", style.Opacity),
			})
		}
	}

	// Validate font size
	if style.FontSize < 0 {
		errors = append(errors, ValidationError{
			Field:   context + ".style.FontSize",
			Message: "font size cannot be negative",
		})
	}

	// Validate stroke width
	if style.StrokeWidth < 0 {
		errors = append(errors, ValidationError{
			Field:   context + ".style.StrokeWidth",
			Message: "stroke width cannot be negative",
		})
	}

	return errors
}
