// Package render provides diagram rendering to various formats.
// This file contains JointJS-based rendering that respects metadata (positions, vertices).
package render

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

//go:embed export.html
var exportHTML embed.FS

// Metadata represents the diagram layout metadata from .d2meta files.
type Metadata struct {
	SourceHash  string                `json:"sourceHash,omitempty"`
	Positions   map[string]NodeOffset `json:"positions,omitempty"`
	Vertices    map[string][]Vertex   `json:"vertices,omitempty"`
	RoutingMode map[string]string     `json:"routingMode,omitempty"`
}

// NodeOffset represents a position offset for a node.
type NodeOffset struct {
	DX float64 `json:"dx"`
	DY float64 `json:"dy"`
}

// Vertex represents a bend point on an edge.
type Vertex struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// RenderResult contains the result of JointJS rendering.
type RenderResult struct {
	Success bool    `json:"success"`
	SVG     string  `json:"svg,omitempty"`
	Width   float64 `json:"width,omitempty"`
	Height  float64 `json:"height,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// RenderWithJointJS renders a diagram using JointJS with metadata support.
// This renders the diagram exactly as it appears in the browser editor,
// including custom positions and edge vertices from the .d2meta file.
//
// Parameters:
//   - d2Svg: The D2-rendered SVG (used as base for node/edge extraction)
//   - metadata: Optional metadata containing positions and vertices
//
// Returns SVG bytes or an error.
func RenderWithJointJS(ctx context.Context, d2Svg []byte, metadata *Metadata) ([]byte, error) {
	// Read the embedded HTML template
	htmlBytes, err := exportHTML.ReadFile("export.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read export template: %w", err)
	}

	// Create data URI for the HTML
	htmlDataURI := "data:text/html;base64," + base64.StdEncoding.EncodeToString(htmlBytes)

	// Prepare metadata JSON
	metadataJSON := "{}"
	if metadata != nil {
		metaBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metaBytes)
	}

	// Escape the SVG for JavaScript
	d2SvgStr := string(d2Svg)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Create headless Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true), // Allow loading external scripts
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	var resultJSON string

	// Execute rendering
	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(htmlDataURI),
		// Wait for the page to be ready (JointJS loaded)
		chromedp.WaitVisible("#jointjs-paper", chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait for exportReady flag
			var ready bool
			for i := 0; i < 50; i++ { // 5 second timeout
				err := chromedp.Evaluate(`window.exportReady === true`, &ready).Do(ctx)
				if err != nil {
					return err
				}
				if ready {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if !ready {
				return fmt.Errorf("timeout waiting for JointJS to load")
			}
			return nil
		}),
		// Call the render function
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				const d2Svg = %s;
				const metadata = %s;
				return JSON.stringify(renderDiagram(d2Svg, metadata));
			})()
		`, jsonString(d2SvgStr), metadataJSON), &resultJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to render with JointJS: %w", err)
	}

	// Parse the result
	var result RenderResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse render result: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("JointJS render failed: %s", result.Error)
	}

	return []byte(result.SVG), nil
}

// RenderWithJointJSToPNG renders using JointJS and converts to PNG.
func RenderWithJointJSToPNG(ctx context.Context, d2Svg []byte, metadata *Metadata, pixelDensity int) ([]byte, error) {
	// First render to SVG with JointJS
	svgBytes, err := RenderWithJointJS(ctx, d2Svg, metadata)
	if err != nil {
		return nil, err
	}

	// Convert to PNG using existing function
	return SVGToPNG(ctx, svgBytes, pixelDensity)
}

// RenderWithJointJSToPDF renders using JointJS and converts to PDF.
func RenderWithJointJSToPDF(ctx context.Context, d2Svg []byte, metadata *Metadata) ([]byte, error) {
	// First render to SVG with JointJS
	svgBytes, err := RenderWithJointJS(ctx, d2Svg, metadata)
	if err != nil {
		return nil, err
	}

	// Convert to PDF using existing function
	return SVGToPDF(ctx, svgBytes)
}

// RenderWithMetadata is a convenience function that renders with metadata if available.
// It falls back to the original D2 SVG if metadata is nil or empty.
func RenderWithMetadata(ctx context.Context, d2Svg []byte, metadata *Metadata, format Format, pixelDensity int) ([]byte, error) {
	// Check if we have meaningful metadata
	hasMetadata := metadata != nil && (len(metadata.Positions) > 0 || len(metadata.Vertices) > 0)

	if !hasMetadata {
		// No metadata, use original SVG
		switch format {
		case FormatSVG:
			return d2Svg, nil
		case FormatPNG:
			return SVGToPNG(ctx, d2Svg, pixelDensity)
		case FormatPDF:
			return SVGToPDF(ctx, d2Svg)
		default:
			return nil, fmt.Errorf("unsupported format: %s", format)
		}
	}

	// Use JointJS rendering with metadata
	switch format {
	case FormatSVG:
		return RenderWithJointJS(ctx, d2Svg, metadata)
	case FormatPNG:
		return RenderWithJointJSToPNG(ctx, d2Svg, metadata, pixelDensity)
	case FormatPDF:
		return RenderWithJointJSToPDF(ctx, d2Svg, metadata)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// jsonString converts a string to a JSON string literal for safe embedding in JavaScript.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
