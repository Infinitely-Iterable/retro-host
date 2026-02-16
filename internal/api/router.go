package api

import (
	"net/http"
	"strings"

	"retro-host/internal/config"
)

func NewRouter(cfg *config.Config, frontendDir, emulatorJSDir string) http.Handler {
	h := NewHandler(cfg)
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/systems", h.GetSystems)
	mux.HandleFunc("/api/roms", h.GetROMs)
	mux.HandleFunc("/api/saves/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetSave(w, r)
		case http.MethodPost:
			h.PostSave(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ROM file serving
	mux.HandleFunc("/roms/", h.ServeROM)

	// EmulatorJS assets
	emulatorFS := http.StripPrefix("/emulatorjs/", http.FileServer(http.Dir(emulatorJSDir)))
	mux.Handle("/emulatorjs/", emulatorFS)

	// Frontend static files (catch-all)
	frontendFS := http.FileServer(http.Dir(frontendDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for root, otherwise serve static files
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, frontendDir+"/index.html")
			return
		}
		if r.URL.Path == "/player.html" || strings.HasPrefix(r.URL.Path, "/css/") || strings.HasPrefix(r.URL.Path, "/js/") {
			frontendFS.ServeHTTP(w, r)
			return
		}
		// Default to index
		http.ServeFile(w, r, frontendDir+"/index.html")
	})

	return mux
}
