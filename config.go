package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AssemblyAIAPIKey string
	UploadDir        string
	Port             string
	MaxFileSize      int64
	AllowedTypes     map[string]bool
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("ASSEMBLYAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ASSEMBLYAI_API_KEY is required")
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	return &Config{
		AssemblyAIAPIKey: apiKey,
		UploadDir:        uploadDir,
		Port:             port,
		MaxFileSize:      100 * 1024 * 1024, // 100MB
		AllowedTypes: map[string]bool{
			".mp3": true,
			".wav": true,
			".m4a": true,
			".mp4": true,
			".avi": true,
			".mov": true,
		},
	}, nil
}
