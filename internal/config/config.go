package config

import "os"

type Config struct {
	ROMDir   string
	DataDir  string
	Port     string
	HostAddr string
}

func Load() *Config {
	return &Config{
		ROMDir:   getEnv("ROM_DIR", "/roms"),
		DataDir:  getEnv("DATA_DIR", "/data"),
		Port:     getEnv("PORT", "8080"),
		HostAddr: getEnv("HOST_ADDR", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
