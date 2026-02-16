package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"retro-host/internal/config"
	"retro-host/internal/scanner"
)

type Handler struct {
	cfg  *config.Config
	tags scanner.TagsConfig
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg:  cfg,
		tags: scanner.LoadTags(cfg.DataDir),
	}
}

type systemInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Core     string `json:"core"`
	ROMCount int    `json:"romCount"`
}

func (h *Handler) GetSystems(w http.ResponseWriter, r *http.Request) {
	roms, err := scanner.ScanROMs(h.cfg.ROMDir, h.tags)
	if err != nil {
		http.Error(w, "failed to scan ROMs", http.StatusInternalServerError)
		return
	}

	var systems []systemInfo
	for _, sys := range scanner.Systems {
		if romList, ok := roms[sys.ID]; ok && len(romList) > 0 {
			systems = append(systems, systemInfo{
				ID:       sys.ID,
				Name:     sys.Name,
				Core:     sys.Core,
				ROMCount: len(romList),
			})
		}
	}

	writeJSON(w, systems)
}

func (h *Handler) GetROMs(w http.ResponseWriter, r *http.Request) {
	systemID := r.URL.Query().Get("system")
	if systemID == "" {
		http.Error(w, "system parameter required", http.StatusBadRequest)
		return
	}

	roms, err := scanner.ScanROMs(h.cfg.ROMDir, h.tags)
	if err != nil {
		http.Error(w, "failed to scan ROMs", http.StatusInternalServerError)
		return
	}

	romList := roms[systemID]
	if romList == nil {
		romList = []scanner.ROM{}
	}

	writeJSON(w, romList)
}

func (h *Handler) ServeROM(w http.ResponseWriter, r *http.Request) {
	// URL: /roms/{system}/{filename}
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/roms/"), "/", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	filename := filepath.Base(parts[1])

	// Find the ROM file — it could be in the root or a subdirectory
	romPath := findROMFile(h.cfg.ROMDir, filename)
	if romPath == "" {
		http.Error(w, "ROM not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, romPath)
}

func findROMFile(romDir, filename string) string {
	var found string
	filepath.WalkDir(romDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() == filename {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func (h *Handler) GetSave(w http.ResponseWriter, r *http.Request) {
	// URL: /api/saves/{system}/{rom}
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/saves/"), "/", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	system, rom := parts[0], parts[1]
	savePath := filepath.Join(h.cfg.DataDir, "saves", system, rom+".sav")

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		http.Error(w, "no save found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, savePath)
}

func (h *Handler) PostSave(w http.ResponseWriter, r *http.Request) {
	// URL: /api/saves/{system}/{rom}
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/saves/"), "/", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	system, rom := parts[0], parts[1]

	// Validate system and rom name to prevent path traversal
	if strings.Contains(system, "..") || strings.Contains(rom, "..") ||
		strings.Contains(system, "/") || strings.Contains(rom, "/") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	saveDir := filepath.Join(h.cfg.DataDir, "saves", system)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		http.Error(w, "failed to create save directory", http.StatusInternalServerError)
		return
	}

	savePath := filepath.Join(saveDir, rom+".sav")

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20)) // 10MB limit
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(savePath, body, 0644); err != nil {
		http.Error(w, "failed to write save", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
