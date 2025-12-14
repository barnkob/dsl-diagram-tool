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

	svg, err := renderD2(r.Context(), req.Source, req.Options)
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

	// Message loop
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case "render":
			svg, err := renderD2(r.Context(), msg.Source, nil)
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
		}
	}
}

// renderD2 renders D2 source to SVG.
func renderD2(ctx context.Context, source string, opts *RenderOptions) ([]byte, error) {
	renderOpts := render.DefaultOptions()

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
