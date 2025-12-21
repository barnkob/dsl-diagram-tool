package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mark/dsl-diagram-tool/pkg/render"
)

// RenderRequest is the request body for POST /api/render.
type RenderRequest struct {
	Source  string         `json:"source"`
	Options *RenderOptions `json:"options,omitempty"`
}

// RenderOptions configures rendering.
type RenderOptions struct {
	ThemeID  int64 `json:"themeId"`
	DarkMode bool  `json:"darkMode"`
	Sketch   bool  `json:"sketch"`
	Padding  int64 `json:"padding"`
}

// RenderResponse is the response body for POST /api/render.
type RenderResponse struct {
	SVG   string `json:"svg,omitempty"`
	Error string `json:"error,omitempty"`
}

// FileResponse is the response body for GET /api/file.
type FileResponse struct {
	Source   string `json:"source"`
	FilePath string `json:"filePath"`
}

// handleRender handles POST /api/render requests.
func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, RenderResponse{Error: "Invalid request body"})
		return
	}

	svg, err := renderD2(r.Context(), req.Source, req.Options, s.C4Mode)
	if err != nil {
		writeJSON(w, http.StatusOK, RenderResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, RenderResponse{SVG: string(svg)})
}

// handleFile handles GET and PUT /api/file requests.
func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleFileGet(w, r)
	case http.MethodPut:
		s.handleFilePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFileGet returns the current file content.
func (s *Server) handleFileGet(w http.ResponseWriter, r *http.Request) {
	if s.FilePath == "" {
		writeJSON(w, http.StatusOK, FileResponse{Source: "", FilePath: ""})
		return
	}

	writeJSON(w, http.StatusOK, FileResponse{
		Source:   s.GetFileContent(),
		FilePath: s.FilePath,
	})
}

// handleFilePut saves content to the file.
func (s *Server) handleFilePut(w http.ResponseWriter, r *http.Request) {
	if s.FilePath == "" {
		http.Error(w, "No file opened", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req struct {
		Source string `json:"source"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update cached content first (prevents file watcher from triggering)
	s.SetFileContent(req.Source)

	// Write to file
	if err := os.WriteFile(s.FilePath, []byte(req.Source), 0644); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"saved": true})
}

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type   string `json:"type"`
	Source string `json:"source,omitempty"`
	SVG    string `json:"svg,omitempty"`
	Error  string `json:"error,omitempty"`

	// Position-related fields
	NodeID    string                `json:"nodeId,omitempty"`    // For position: node identifier
	DX        float64               `json:"dx,omitempty"`        // For position: x offset
	DY        float64               `json:"dy,omitempty"`        // For position: y offset
	Positions map[string]NodeOffset `json:"positions,omitempty"` // For positions: all offsets

	// Vertex-related fields
	EdgeID      string              `json:"edgeId,omitempty"`      // For vertices: edge identifier
	Vertices    []Vertex            `json:"vertices,omitempty"`    // For vertices: single edge vertices
	AllVertices map[string][]Vertex `json:"allVertices,omitempty"` // For vertices: all edge vertices

	// Routing mode fields
	RoutingMode    string            `json:"routingMode,omitempty"`    // For routing: edge routing mode (direct, orthogonal)
	AllRoutingMode map[string]string `json:"allRoutingMode,omitempty"` // For positions: all routing modes

	// Label position fields
	LabelDistance      float64                  `json:"labelDistance,omitempty"`      // For label-position: distance along edge (0-1)
	LabelOffsetX       float64                  `json:"labelOffsetX,omitempty"`       // For label-position: X offset
	LabelOffsetY       float64                  `json:"labelOffsetY,omitempty"`       // For label-position: Y offset
	AllLabelPositions  map[string]LabelPosition `json:"allLabelPositions,omitempty"`  // For positions: all label positions
}

// handleWebSocket handles WebSocket connections.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Register client
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	// Send initial file content
	if s.FilePath != "" {
		conn.WriteJSON(WSMessage{
			Type:   "file-changed",
			Source: s.GetFileContent(),
		})
	}

	// Send initial positions, vertices, routing modes, and label positions
	meta := s.GetMetadata()
	if meta.HasPositions() || meta.HasVertices() || meta.HasRoutingModes() || meta.HasLabelPositions() {
		conn.WriteJSON(WSMessage{
			Type:              "positions",
			Positions:         meta.Positions,
			AllVertices:       meta.Vertices,
			AllRoutingMode:    meta.RoutingMode,
			AllLabelPositions: meta.LabelPositions,
		})
	}

	// Message loop
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "render":
			svg, err := renderD2(r.Context(), msg.Source, nil, s.C4Mode)
			if err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: err.Error(),
				})
			} else {
				conn.WriteJSON(WSMessage{
					Type: "rendered",
					SVG:  string(svg),
				})
			}

		case "save":
			if s.FilePath == "" {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "No file opened",
				})
				continue
			}

			// Update cached content
			s.SetFileContent(msg.Source)

			// Write to file
			if err := os.WriteFile(s.FilePath, []byte(msg.Source), 0644); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to save file",
				})
			} else {
				conn.WriteJSON(WSMessage{Type: "saved"})
			}

		case "position":
			// Update node position
			if msg.NodeID == "" {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "nodeId is required",
				})
				continue
			}

			if err := s.SetNodePosition(msg.NodeID, msg.DX, msg.DY); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to save position: " + err.Error(),
				})
				continue
			}

			// Acknowledge position saved
			conn.WriteJSON(WSMessage{
				Type:   "position-saved",
				NodeID: msg.NodeID,
			})

		case "vertices":
			// Update edge vertices
			if msg.EdgeID == "" {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "edgeId is required",
				})
				continue
			}

			if err := s.SetEdgeVertices(msg.EdgeID, msg.Vertices); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to save vertices: " + err.Error(),
				})
				continue
			}

			// Acknowledge vertices saved
			conn.WriteJSON(WSMessage{
				Type:   "vertices-saved",
				EdgeID: msg.EdgeID,
			})

		case "routing":
			// Update edge routing mode
			if msg.EdgeID == "" {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "edgeId is required",
				})
				continue
			}

			if err := s.SetRoutingMode(msg.EdgeID, msg.RoutingMode); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to save routing mode: " + err.Error(),
				})
				continue
			}

			// Acknowledge routing mode saved
			conn.WriteJSON(WSMessage{
				Type:        "routing-saved",
				EdgeID:      msg.EdgeID,
				RoutingMode: msg.RoutingMode,
			})

		case "label-position":
			// Update edge label position
			if msg.EdgeID == "" {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "edgeId is required",
				})
				continue
			}

			if err := s.SetLabelPosition(msg.EdgeID, msg.LabelDistance, msg.LabelOffsetX, msg.LabelOffsetY); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to save label position: " + err.Error(),
				})
				continue
			}

			// Acknowledge label position saved
			conn.WriteJSON(WSMessage{
				Type:   "label-position-saved",
				EdgeID: msg.EdgeID,
			})

		case "clear-positions":
			// Clear all positions and vertices
			if err := s.ClearAllPositions(); err != nil {
				conn.WriteJSON(WSMessage{
					Type:  "error",
					Error: "Failed to clear positions: " + err.Error(),
				})
				continue
			}

			// Broadcast to all clients
			s.broadcast(WSMessage{
				Type: "positions-cleared",
			})
		}
	}
}

// renderD2 renders D2 source to SVG.
func renderD2(ctx context.Context, source string, opts *RenderOptions, c4Mode bool) ([]byte, error) {
	renderOpts := render.DefaultOptions()

	// Apply C4 mode defaults (Terminal theme + inject C4 classes)
	if c4Mode {
		renderOpts.ThemeID = 8
		// Prepend C4 class definitions to the source
		source = render.ApplyC4Theme(source)
	}

	// Allow explicit options to override C4 defaults
	if opts != nil {
		if opts.ThemeID != 0 {
			renderOpts.ThemeID = opts.ThemeID
		}
		renderOpts.DarkMode = opts.DarkMode
		renderOpts.Sketch = opts.Sketch
		if opts.Padding != 0 {
			renderOpts.Padding = opts.Padding
		}
	}

	// Use a timeout for rendering
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return render.RenderFromSource(ctx, source, renderOpts)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
