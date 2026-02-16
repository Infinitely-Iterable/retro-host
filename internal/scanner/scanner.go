package scanner

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type System struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Extensions []string `json:"-"`
	Core       string   `json:"core"`
}

type ROM struct {
	Name     string `json:"name"`
	FileName string `json:"fileName"`
	System   string `json:"system"`
	Tag      string `json:"tag"`
}

var Systems = []System{
	{ID: "gb", Name: "Game Boy", Extensions: []string{".gb"}, Core: "gb"},
	{ID: "gbc", Name: "Game Boy Color", Extensions: []string{".gbc"}, Core: "gb"},
	{ID: "gba", Name: "Game Boy Advance", Extensions: []string{".gba"}, Core: "vba_next"},
	{ID: "nes", Name: "NES", Extensions: []string{".nes"}, Core: "nes"},
	{ID: "snes", Name: "SNES", Extensions: []string{".smc", ".sfc"}, Core: "snes"},
}

// File extensions to ignore (save files, patches, etc.)
var ignoredExtensions = map[string]bool{
	".srm": true,
	".sav": true,
	".bak": true,
	".ips": true,
	".ups": true,
}

var extToSystem map[string]*System
var dirToSystem map[string]*System

func init() {
	extToSystem = make(map[string]*System)
	dirToSystem = make(map[string]*System)
	for i := range Systems {
		for _, ext := range Systems[i].Extensions {
			extToSystem[ext] = &Systems[i]
		}
		dirToSystem[Systems[i].ID] = &Systems[i]
	}
}

func SystemByID(id string) *System {
	for i := range Systems {
		if Systems[i].ID == id {
			return &Systems[i]
		}
	}
	return nil
}

// TagsConfig maps ROM filenames to tags.
// Example: {"Pokemon Fire Red.gba": "RPG", "Super Mario World.smc": "Platformer"}
type TagsConfig map[string]string

func LoadTags(dataDir string) TagsConfig {
	tagsPath := filepath.Join(dataDir, "tags.json")
	data, err := os.ReadFile(tagsPath)
	if err != nil {
		return TagsConfig{}
	}
	var tags TagsConfig
	if err := json.Unmarshal(data, &tags); err != nil {
		log.Printf("Warning: failed to parse tags.json: %v", err)
		return TagsConfig{}
	}
	return tags
}

func ScanROMs(romDir string, tags TagsConfig) (map[string][]ROM, error) {
	result := make(map[string][]ROM)

	err := filepath.WalkDir(romDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Skip ignored file types
		if ignoredExtensions[ext] {
			return nil
		}

		// First check: is this file inside a system-named directory?
		var sys *System
		rel, _ := filepath.Rel(romDir, path)
		parentDir := strings.ToLower(strings.SplitN(rel, string(filepath.Separator), 2)[0])
		if dirSys, ok := dirToSystem[parentDir]; ok && parentDir != rel {
			sys = dirSys
		}

		// Fallback: detect by file extension
		if sys == nil {
			extSys, ok := extToSystem[ext]
			if !ok {
				return nil
			}
			sys = extSys
		}

		name := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		tag := tags[d.Name()]

		result[sys.ID] = append(result[sys.ID], ROM{
			Name:     name,
			FileName: d.Name(),
			System:   sys.ID,
			Tag:      tag,
		})

		return nil
	})

	return result, err
}
