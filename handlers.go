package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
)

func (app *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	// Serve static files from the frontend directory
	if strings.HasPrefix(r.URL.Path, "/static/") ||
		strings.HasSuffix(r.URL.Path, ".js") ||
		strings.HasSuffix(r.URL.Path, ".css") ||
		strings.HasSuffix(r.URL.Path, ".ico") {
		fs := http.FileServer(http.Dir(app.config.Server.StaticDir))
		fs.ServeHTTP(w, r)
		return
	}

	// For all other paths, serve the main index.html
	if r.URL.Path != "/" {
		// Redirect to root for SPA routing
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	http.ServeFile(w, r, filepath.Join(app.config.Server.StaticDir, "index.html"))
}

func (app *App) apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticateRequest(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"authenticated": false}`))
		return
	}

	response := map[string]interface{}{
		"authenticated": true,
		"role":          user.Role,
		"username":      user.Username,
		"user_id":       user.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
