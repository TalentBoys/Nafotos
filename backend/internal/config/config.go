package config

import (
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Port          string
	DBPath        string
	ConfigDir     string
	UploadDir     string
	ThumbsDir     string
	MountedDirs   []string
	AllowedOrigin string
}

func Load() *Config {
	configDir := getEnv("CONFIG_DIR", "/config")
	uploadDir := getEnv("UPLOAD_DIR", "/upload")

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		ConfigDir:     configDir,
		UploadDir:     uploadDir,
		DBPath:        filepath.Join(configDir, "awesome-sharing.db"),
		ThumbsDir:     filepath.Join(configDir, "thumbs"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),
		MountedDirs:   []string{configDir, uploadDir},
	}

	// Ensure all required directories exist
	if err := os.MkdirAll(cfg.ConfigDir, 0755); err != nil {
		log.Printf("Warning: could not create config directory: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Printf("Warning: could not create upload directory: %v", err)
	}
	if err := os.MkdirAll(cfg.ThumbsDir, 0755); err != nil {
		log.Printf("Warning: could not create thumbs directory: %v", err)
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
