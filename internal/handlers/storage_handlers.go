package handlers

import (
	"github.com/go-chi/chi"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetImage returns the image from the application storage.
func (m *Repository) GetImage(w http.ResponseWriter, r *http.Request) {
	basePath := "storage/images"

	filePath := chi.URLParam(r, "*")

	fullPath := filepath.Join(basePath, filePath)

	if !strings.HasPrefix(fullPath, filepath.Clean(basePath)+string(os.PathSeparator)) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(fullPath); err == nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		w.Header().Set("Expires", "0")                                         // Proxies.

		http.ServeFile(w, r, fullPath)
	} else {
		http.NotFound(w, r)
	}
}
