package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// Metadata stores position overrides for diagram nodes and edge vertices.
// Stored in a .d2meta file alongside the .d2 file.
type Metadata struct {
	Version        int                      `json:"version"`
	Positions      map[string]NodeOffset    `json:"positions"`
	Vertices       map[string][]Vertex      `json:"vertices,omitempty"`
	RoutingMode    map[string]string        `json:"routingMode,omitempty"`
	LabelPositions map[string]LabelPosition `json:"labelPositions,omitempty"`
	SourceHash     string                   `json:"sourceHash"`
}

// NormalizeEdgeID ensures consistent edge ID format by decoding HTML entities.
// Edge IDs may contain HTML-encoded arrows (e.g., "-&gt;" instead of "->").
func NormalizeEdgeID(edgeID string) string {
	return html.UnescapeString(edgeID)
}

// NodeOffset represents the offset from auto-layout position.
type NodeOffset struct {
	DX float64 `json:"dx"`
	DY float64 `json:"dy"`
}

// Vertex represents a bend point coordinate on an edge.
type Vertex struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// LabelPosition represents the position of an edge label.
// Distance is 0-1 along the edge path, Offset is perpendicular displacement.
type LabelPosition struct {
	Distance float64 `json:"distance"`
	OffsetX  float64 `json:"offsetX,omitempty"`
	OffsetY  float64 `json:"offsetY,omitempty"`
}

// NewMetadata creates a new empty metadata structure.
func NewMetadata() *Metadata {
	return &Metadata{
		Version:        1,
		Positions:      make(map[string]NodeOffset),
		Vertices:       make(map[string][]Vertex),
		RoutingMode:    make(map[string]string),
		LabelPositions: make(map[string]LabelPosition),
	}
}

// MetadataPath returns the .d2meta path for a given .d2 file path.
func MetadataPath(d2Path string) string {
	ext := filepath.Ext(d2Path)
	return strings.TrimSuffix(d2Path, ext) + ".d2meta"
}

// LoadMetadata loads metadata from the .d2meta file.
// Returns empty metadata if file doesn't exist.
func LoadMetadata(d2Path string) (*Metadata, error) {
	metaPath := MetadataPath(d2Path)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewMetadata(), nil
		}
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	if meta.Positions == nil {
		meta.Positions = make(map[string]NodeOffset)
	}
	if meta.Vertices == nil {
		meta.Vertices = make(map[string][]Vertex)
	}
	if meta.RoutingMode == nil {
		meta.RoutingMode = make(map[string]string)
	}
	if meta.LabelPositions == nil {
		meta.LabelPositions = make(map[string]LabelPosition)
	}

	return &meta, nil
}

// SaveMetadata saves metadata to the .d2meta file.
func SaveMetadata(d2Path string, meta *Metadata) error {
	metaPath := MetadataPath(d2Path)

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0644)
}

// HashSource computes a SHA256 hash of the D2 source content.
func HashSource(source string) string {
	hash := sha256.Sum256([]byte(source))
	return hex.EncodeToString(hash[:8]) // First 8 bytes is enough
}

// ValidateAndClean checks if source hash matches and clears positions/vertices/routing if not.
// Returns true if data was cleared.
func (m *Metadata) ValidateAndClean(currentSource string) bool {
	currentHash := HashSource(currentSource)

	if m.SourceHash != currentHash {
		// Source changed, clear all positions, vertices, routing modes, and label positions
		m.Positions = make(map[string]NodeOffset)
		m.Vertices = make(map[string][]Vertex)
		m.RoutingMode = make(map[string]string)
		m.LabelPositions = make(map[string]LabelPosition)
		m.SourceHash = currentHash
		return true
	}

	return false
}

// SetPosition updates or adds a position offset for a node.
func (m *Metadata) SetPosition(nodeID string, dx, dy float64) {
	m.Positions[nodeID] = NodeOffset{DX: dx, DY: dy}
}

// GetPosition returns the position offset for a node.
// Returns zero offset if not found.
func (m *Metadata) GetPosition(nodeID string) NodeOffset {
	if offset, ok := m.Positions[nodeID]; ok {
		return offset
	}
	return NodeOffset{}
}

// ClearPosition removes the position offset for a node.
func (m *Metadata) ClearPosition(nodeID string) {
	delete(m.Positions, nodeID)
}

// HasPositions returns true if there are any position overrides.
func (m *Metadata) HasPositions() bool {
	return len(m.Positions) > 0
}

// SetVertices updates or adds vertices for an edge.
// Edge IDs are normalized to handle HTML entity encoding.
func (m *Metadata) SetVertices(edgeID string, vertices []Vertex) {
	normalizedID := NormalizeEdgeID(edgeID)
	if len(vertices) == 0 {
		delete(m.Vertices, normalizedID)
	} else {
		m.Vertices[normalizedID] = vertices
	}
}

// GetVertices returns the vertices for an edge.
// Returns empty slice if not found.
func (m *Metadata) GetVertices(edgeID string) []Vertex {
	normalizedID := NormalizeEdgeID(edgeID)
	if vertices, ok := m.Vertices[normalizedID]; ok {
		return vertices
	}
	return []Vertex{}
}

// HasVertices returns true if there are any edge vertices.
func (m *Metadata) HasVertices() bool {
	return len(m.Vertices) > 0
}

// SetRoutingMode updates or adds a routing mode for an edge.
// Valid modes are "direct" (default), "orthogonal".
func (m *Metadata) SetRoutingMode(edgeID string, mode string) {
	normalizedID := NormalizeEdgeID(edgeID)
	if mode == "" || mode == "direct" {
		delete(m.RoutingMode, normalizedID) // direct is default, no need to store
	} else {
		m.RoutingMode[normalizedID] = mode
	}
}

// GetRoutingMode returns the routing mode for an edge.
// Returns "direct" if not found.
func (m *Metadata) GetRoutingMode(edgeID string) string {
	normalizedID := NormalizeEdgeID(edgeID)
	if mode, ok := m.RoutingMode[normalizedID]; ok {
		return mode
	}
	return "direct"
}

// HasRoutingModes returns true if there are any non-default routing modes.
func (m *Metadata) HasRoutingModes() bool {
	return len(m.RoutingMode) > 0
}

// SetLabelPosition updates or adds a label position for an edge.
func (m *Metadata) SetLabelPosition(edgeID string, distance, offsetX, offsetY float64) {
	normalizedID := NormalizeEdgeID(edgeID)
	// Only store if not default (0.5, 0, 0)
	if distance == 0.5 && offsetX == 0 && offsetY == 0 {
		delete(m.LabelPositions, normalizedID)
	} else {
		m.LabelPositions[normalizedID] = LabelPosition{
			Distance: distance,
			OffsetX:  offsetX,
			OffsetY:  offsetY,
		}
	}
}

// GetLabelPosition returns the label position for an edge.
// Returns default position (0.5, 0, 0) if not found.
func (m *Metadata) GetLabelPosition(edgeID string) LabelPosition {
	normalizedID := NormalizeEdgeID(edgeID)
	if pos, ok := m.LabelPositions[normalizedID]; ok {
		return pos
	}
	return LabelPosition{Distance: 0.5}
}

// HasLabelPositions returns true if there are any custom label positions.
func (m *Metadata) HasLabelPositions() bool {
	return len(m.LabelPositions) > 0
}
