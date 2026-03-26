package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup logging: stdout + file
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		slog.Error("failed to create upload dir", "error", err)
		os.Exit(1)
	}

	// Setup client and handlers
	client := NewAssemblyAIClient(cfg.AssemblyAIAPIKey)
	h := NewHandler(cfg, client)

	// Router
	r := chi.NewRouter()
	r.Use(CORSMiddleware)
	r.Get("/health", h.Health)
	r.Options("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r.Post("/upload", h.Upload)

	slog.Info("starting server", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
