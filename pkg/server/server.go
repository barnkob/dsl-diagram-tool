// Package server provides the HTTP server for the diagram editor web interface.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

// Server represents the diagram editor HTTP server.
type Server struct {
	// Configuration
	Port     int
	FilePath string // Path to the D2 file being edited

	// Internal state
	httpServer *http.Server
	watcher    *fsnotify.Watcher
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	upgrader   websocket.Upgrader

	// Current file content (cached)
	fileContent   string
	fileContentMu sync.RWMutex

	// Position metadata
	metadata   *Metadata
	metadataMu sync.RWMutex
}

// Options configures the server.
type Options struct {
	Port     int
	FilePath string
	DevMode  bool // If true, serve from filesystem instead of embedded
}

// New creates a new server instance.
func New(opts Options) (*Server, error) {
	if opts.Port == 0 {
		opts.Port = 8080
	}

	s := &Server{
		Port:     opts.Port,
		FilePath: opts.FilePath,
		clients:  make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for local development
			},
		},
	}

	// Load initial file content if file specified
	if opts.FilePath != "" {
		absPath, err := filepath.Abs(opts.FilePath)
		if err != nil {
			return nil, fmt.Errorf("invalid file path: %w", err)
		}
		s.FilePath = absPath

		content, err := os.ReadFile(s.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		s.fileContent = string(content)

		// Load metadata (positions)
		meta, err := LoadMetadata(s.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load metadata: %w", err)
		}
		s.metadata = meta

		// Validate metadata against current source
		if s.metadata.ValidateAndClean(s.fileContent) {
			// Source changed, save cleared metadata
			_ = SaveMetadata(s.FilePath, s.metadata)
		}
	} else {
		s.metadata = NewMetadata()
	}

	return s, nil
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/render", s.handleRender)
	mux.HandleFunc("/api/file", s.handleFile)
	mux.HandleFunc("/api/ws", s.handleWebSocket)

	// Static files (frontend)
	mux.HandleFunc("/", s.handleStatic)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	// Start file watcher if we have a file
	if s.FilePath != "" {
		if err := s.startFileWatcher(); err != nil {
			return fmt.Errorf("failed to start file watcher: %w", err)
		}
	}

	// Start server
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	// Stop file watcher
	if s.watcher != nil {
		s.watcher.Close()
	}

	// Close all WebSocket connections
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMu.Unlock()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// startFileWatcher starts watching the D2 file for external changes.
func (s *Server) startFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = watcher

	// Watch the directory (more reliable for file saves)
	dir := filepath.Dir(s.FilePath)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	go s.watchFileChanges()
	return nil
}

// watchFileChanges handles file system events.
func (s *Server) watchFileChanges() {
	// Debounce timer
	var debounceTimer *time.Timer
	debounceDelay := 100 * time.Millisecond

	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			// Only care about our file
			if filepath.Clean(event.Name) != filepath.Clean(s.FilePath) {
				continue
			}

			// Only care about write events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				s.handleFileChanged()
			})

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "file watcher error: %v\n", err)
		}
	}
}

// handleFileChanged is called when the D2 file changes externally.
func (s *Server) handleFileChanged() {
	content, err := os.ReadFile(s.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read changed file: %v\n", err)
		return
	}

	newContent := string(content)

	// Check if content actually changed
	s.fileContentMu.RLock()
	changed := newContent != s.fileContent
	s.fileContentMu.RUnlock()

	if !changed {
		return
	}

	// Update cached content
	s.fileContentMu.Lock()
	s.fileContent = newContent
	s.fileContentMu.Unlock()

	// Check if positions should be cleared (source hash changed)
	s.metadataMu.Lock()
	positionsCleared := s.metadata.ValidateAndClean(newContent)
	if positionsCleared {
		_ = SaveMetadata(s.FilePath, s.metadata)
	}
	s.metadataMu.Unlock()

	// Broadcast file change to all clients
	s.broadcast(WSMessage{
		Type:   "file-changed",
		Source: newContent,
	})

	// If positions were cleared, notify clients
	if positionsCleared {
		s.broadcast(WSMessage{
			Type: "positions-cleared",
		})
	}
}

// broadcast sends a message to all connected WebSocket clients.
func (s *Server) broadcast(msg WSMessage) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for conn := range s.clients {
		if err := conn.WriteJSON(msg); err != nil {
			// Connection will be cleaned up by read loop
			continue
		}
	}
}

// GetFileContent returns the current file content.
func (s *Server) GetFileContent() string {
	s.fileContentMu.RLock()
	defer s.fileContentMu.RUnlock()
	return s.fileContent
}

// SetFileContent updates the cached file content.
func (s *Server) SetFileContent(content string) {
	s.fileContentMu.Lock()
	s.fileContent = content
	s.fileContentMu.Unlock()
}

// GetMetadata returns a copy of the current metadata.
func (s *Server) GetMetadata() *Metadata {
	s.metadataMu.RLock()
	defer s.metadataMu.RUnlock()

	// Return a copy to avoid race conditions
	copy := &Metadata{
		Version:    s.metadata.Version,
		SourceHash: s.metadata.SourceHash,
		Positions:  make(map[string]NodeOffset),
	}
	for k, v := range s.metadata.Positions {
		copy.Positions[k] = v
	}
	return copy
}

// SetNodePosition updates a node's position offset and saves metadata.
func (s *Server) SetNodePosition(nodeID string, dx, dy float64) error {
	s.metadataMu.Lock()
	s.metadata.SetPosition(nodeID, dx, dy)
	s.metadataMu.Unlock()

	if s.FilePath != "" {
		return SaveMetadata(s.FilePath, s.metadata)
	}
	return nil
}

// ClearAllPositions clears all position overrides.
func (s *Server) ClearAllPositions() error {
	s.metadataMu.Lock()
	s.metadata.Positions = make(map[string]NodeOffset)
	s.metadataMu.Unlock()

	if s.FilePath != "" {
		return SaveMetadata(s.FilePath, s.metadata)
	}
	return nil
}
