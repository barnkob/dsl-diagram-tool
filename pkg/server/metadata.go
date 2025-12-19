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

// Metadata stores position overrides for diagram nodes and edge waypoints.
// Stored in a .d2meta file alongside the .d2 file.
type Metadata struct {
	Version     int                      `json:"version"`
	Positions   map[string]NodeOffset    `json:"positions"`
	Waypoints   map[string][]EdgePoint   `json:"waypoints,omitempty"`
	RoutingMode map[string]string        `json:"routingMode,omitempty"`
	SourceHash  string                   `json:"sourceHash"`
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

// EdgePoint represents a waypoint coordinate on an edge.
type EdgePoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NewMetadata creates a new empty metadata structure.
func NewMetadata() *Metadata {
	return &Metadata{
		Version:     1,
		Positions:   make(map[string]NodeOffset),
		Waypoints:   make(map[string][]EdgePoint),
		RoutingMode: make(map[string]string),
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
	if meta.Waypoints == nil {
		meta.Waypoints = make(map[string][]EdgePoint)
	}
	if meta.RoutingMode == nil {
		meta.RoutingMode = make(map[string]string)
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

// ValidateAndClean checks if source hash matches and clears positions/waypoints/routing if not.
// Returns true if data was cleared.
func (m *Metadata) ValidateAndClean(currentSource string) bool {
	currentHash := HashSource(currentSource)

	if m.SourceHash != currentHash {
		// Source changed, clear all positions, waypoints, and routing modes
		m.Positions = make(map[string]NodeOffset)
		m.Waypoints = make(map[string][]EdgePoint)
		m.RoutingMode = make(map[string]string)
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

// SetWaypoints updates or adds waypoints for an edge.
// Edge IDs are normalized to handle HTML entity encoding.
func (m *Metadata) SetWaypoints(edgeID string, waypoints []EdgePoint) {
	normalizedID := NormalizeEdgeID(edgeID)
	if len(waypoints) == 0 {
		delete(m.Waypoints, normalizedID)
	} else {
		m.Waypoints[normalizedID] = waypoints
	}
}

// GetWaypoints returns the waypoints for an edge.
// Returns empty slice if not found.
func (m *Metadata) GetWaypoints(edgeID string) []EdgePoint {
	normalizedID := NormalizeEdgeID(edgeID)
	if waypoints, ok := m.Waypoints[normalizedID]; ok {
		return waypoints
	}
	return []EdgePoint{}
}

// HasWaypoints returns true if there are any edge waypoints.
func (m *Metadata) HasWaypoints() bool {
	return len(m.Waypoints) > 0
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
