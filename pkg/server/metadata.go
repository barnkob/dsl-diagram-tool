package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Metadata stores position overrides for diagram nodes.
// Stored in a .d2meta file alongside the .d2 file.
type Metadata struct {
	Version    int                    `json:"version"`
	Positions  map[string]NodeOffset  `json:"positions"`
	SourceHash string                 `json:"sourceHash"`
}

// NodeOffset represents the offset from auto-layout position.
type NodeOffset struct {
	DX float64 `json:"dx"`
	DY float64 `json:"dy"`
}

// NewMetadata creates a new empty metadata structure.
func NewMetadata() *Metadata {
	return &Metadata{
		Version:   1,
		Positions: make(map[string]NodeOffset),
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

// ValidateAndClean checks if source hash matches and clears positions if not.
// Returns true if positions were cleared.
func (m *Metadata) ValidateAndClean(currentSource string) bool {
	currentHash := HashSource(currentSource)

	if m.SourceHash != currentHash {
		// Source changed, clear all positions
		m.Positions = make(map[string]NodeOffset)
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
