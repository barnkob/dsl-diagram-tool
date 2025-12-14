package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web/dist/*
var embeddedFS embed.FS

// handleStatic serves static files from the embedded filesystem.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Get the path
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Try to serve from embedded files
	subFS, err := fs.Sub(embeddedFS, "web/dist")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Remove leading slash for fs.FS
	filePath := strings.TrimPrefix(path, "/")

	// Try to open the file
	file, err := subFS.Open(filePath)
	if err != nil {
		// If not found, serve index.html for SPA routing
		if filePath != "index.html" {
			file, err = subFS.Open("index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Set content type based on extension
	contentType := getContentType(filePath)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Serve the file
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(readSeeker))
}

// readSeeker combines io.Reader and io.Seeker
type readSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

// getContentType returns the content type for a file extension.
func getContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return "application/json; charset=utf-8"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	default:
		return ""
	}
}
