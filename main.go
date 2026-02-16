package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"retro-host/internal/api"
	"retro-host/internal/config"
	"retro-host/internal/scanner"
)

func main() {
	cfg := config.Load()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			cmdList(cfg)
			return
		case "covers":
			cmdCovers(cfg)
			return
		case "play":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: retro-host play <rom-name>\n")
				fmt.Fprintf(os.Stderr, "  Use 'retro-host list' to see available ROMs\n")
				os.Exit(1)
			}
			cmdPlay(cfg, strings.Join(os.Args[2:], " "))
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	cmdServe(cfg)
}

func printUsage() {
	fmt.Println("RetroHost - Self-hosted retro gaming platform")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  retro-host              Start the web server")
	fmt.Println("  retro-host list         List all available ROMs")
	fmt.Println("  retro-host covers       Show cover art status for all ROMs")
	fmt.Println("  retro-host play <name>  Print URL to play a ROM")
	fmt.Println("  retro-host help         Show this help message")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  ROM_DIR    Directory containing ROM files (default: /roms)")
	fmt.Println("  DATA_DIR   Directory for save data (default: /data)")
	fmt.Println("  PORT       Server port (default: 8080)")
	fmt.Println("  HOST_ADDR  External address for URLs (default: localhost:PORT)")
}

func cmdServe(cfg *config.Config) {
	frontendDir := findDir("frontend", "./frontend", "/app/frontend")
	emulatorJSDir := findDir("emulatorjs", "./emulatorjs", "/app/emulatorjs")

	router := api.NewRouter(cfg, frontendDir, emulatorJSDir)

	addr := ":" + cfg.Port
	log.Printf("RetroHost starting on %s", addr)
	log.Printf("ROM directory: %s", cfg.ROMDir)
	log.Printf("Data directory: %s", cfg.DataDir)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}

func cmdList(cfg *config.Config) {
	tags := scanner.LoadTags(cfg.DataDir)
	roms, err := scanner.ScanROMs(cfg.ROMDir, tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning ROMs: %v\n", err)
		os.Exit(1)
	}

	if len(roms) == 0 {
		fmt.Println("No ROMs found in", cfg.ROMDir)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SYSTEM\tROM\tFILE")
	fmt.Fprintln(w, "------\t---\t----")

	for _, sys := range scanner.Systems {
		romList, ok := roms[sys.ID]
		if !ok {
			continue
		}
		sort.Slice(romList, func(i, j int) bool {
			return romList[i].Name < romList[j].Name
		})
		for _, rom := range romList {
			fmt.Fprintf(w, "%s\t%s\t%s\n", sys.Name, rom.Name, rom.FileName)
		}
	}
	w.Flush()
}

func cmdPlay(cfg *config.Config, query string) {
	tags := scanner.LoadTags(cfg.DataDir)
	roms, err := scanner.ScanROMs(cfg.ROMDir, tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning ROMs: %v\n", err)
		os.Exit(1)
	}

	query = strings.ToLower(query)
	var matches []scanner.ROM

	for _, romList := range roms {
		for _, rom := range romList {
			if strings.Contains(strings.ToLower(rom.Name), query) ||
				strings.Contains(strings.ToLower(rom.FileName), query) {
				matches = append(matches, rom)
			}
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "No ROM found matching '%s'\n", query)
		fmt.Fprintf(os.Stderr, "Use 'retro-host list' to see available ROMs\n")
		os.Exit(1)
	}

	host := cfg.HostAddr
	if host == "" {
		host = "localhost:" + cfg.Port
	}

	if len(matches) == 1 {
		rom := matches[0]
		sys := scanner.SystemByID(rom.System)
		fmt.Printf("Open this URL to play %s:\n\n", rom.Name)
		fmt.Printf("  http://%s/player.html?system=%s&rom=%s&core=%s\n\n",
			host, rom.System, rom.FileName, sys.Core)
		return
	}

	fmt.Printf("Multiple ROMs match '%s':\n\n", query)
	for _, rom := range matches {
		sys := scanner.SystemByID(rom.System)
		fmt.Printf("  [%s] %s\n    http://%s/player.html?system=%s&rom=%s&core=%s\n\n",
			sys.Name, rom.Name, host, rom.System, rom.FileName, sys.Core)
	}
}

func cmdCovers(cfg *config.Config) {
	tags := scanner.LoadTags(cfg.DataDir)
	roms, err := scanner.ScanROMs(cfg.ROMDir, tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning ROMs: %v\n", err)
		os.Exit(1)
	}

	if len(roms) == 0 {
		fmt.Println("No ROMs found in", cfg.ROMDir)
		return
	}

	statuses := scanner.CheckCovers(cfg.DataDir, roms)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SYSTEM\tROM\tCOVER")
	fmt.Fprintln(w, "------\t---\t-----")

	missing := 0
	for _, s := range statuses {
		status := "✓"
		if !s.HasCover {
			status = "✗ missing"
			missing++
		}
		sys := scanner.SystemByID(s.ROM.System)
		sysName := s.ROM.System
		if sys != nil {
			sysName = sys.Name
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", sysName, s.ROM.Name, status)
	}
	w.Flush()

	fmt.Printf("\n%d/%d ROMs have covers", len(statuses)-missing, len(statuses))
	if missing > 0 {
		fmt.Printf(" (%d missing)", missing)
	}
	fmt.Println()
	fmt.Printf("\nPlace cover images in: %s/covers/<system>/<rom-name>.{png,jpg,webp}\n", cfg.DataDir)
}

func findDir(name, localPath, containerPath string) string {
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		return localPath
	}
	if info, err := os.Stat(containerPath); err == nil && info.IsDir() {
		return containerPath
	}
	log.Printf("Warning: %s directory not found at %s or %s", name, localPath, containerPath)
	return localPath
}
