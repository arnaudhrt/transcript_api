package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Handler struct {
	cfg    *Config
	client *AssemblyAIClient
}

func NewHandler(cfg *Config, client *AssemblyAIClient) *Handler {
	return &Handler{cfg: cfg, client: client}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "healthy",
		Message: "Meeting Transcription API is running",
	})
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parse multipart form (limit to MaxFileSize + 1MB overhead for headers)
	if err := r.ParseMultipartForm(h.cfg.MaxFileSize + 1<<20); err != nil {
		slog.Error("failed to parse multipart form", "error", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "Invalid multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("no file in request", "error", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "No file provided"})
		return
	}
	defer file.Close()

	// Validate
	if err := ValidateFile(header.Filename, header.Size, h.cfg); err != nil {
		if ve, ok := err.(*ValidationError); ok {
			writeJSON(w, ve.Code, ErrorResponse{Detail: ve.Message})
			return
		}
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: err.Error()})
		return
	}

	// Optional speakers_expected
	var speakersExpected *int
	if s := r.FormValue("speakers_expected"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Detail: "speakers_expected must be a positive integer"})
			return
		}
		speakersExpected = &n
	}

	slog.Info("processing file", "filename", header.Filename, "size", header.Size)

	// Write to temp file
	ext := filepath.Ext(header.Filename)
	tmpFile, err := os.CreateTemp(h.cfg.UploadDir, "upload-*"+ext)
	if err != nil {
		slog.Error("failed to create temp file", "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Detail: "Failed to process file"})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		slog.Error("failed to write temp file", "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Detail: "Failed to process file"})
		return
	}
	tmpFile.Close()

	// Transcribe
	result, err := h.client.Transcribe(r.Context(), tmpFile.Name(), speakersExpected)
	if err != nil {
		slog.Error("transcription failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Detail: "Transcription failed: " + err.Error()})
		return
	}

	// Build response
	resp := buildTranscriptResponse(result, time.Since(start).Seconds())

	slog.Info("transcription complete",
		"processing_time", resp.ProcessingTime,
		"segments", len(resp.Segments),
	)

	writeJSON(w, http.StatusOK, resp)
}

func buildTranscriptResponse(t *TranscriptStatus, processingTime float64) TranscriptResponse {
	transcript := ""
	if t.Text != nil {
		transcript = *t.Text
	}

	language := "unknown"
	if t.LanguageCode != nil {
		language = *t.LanguageCode
	}

	segments := make([]Segment, 0, len(t.Utterances))
	if len(t.Utterances) > 0 {
		for _, u := range t.Utterances {
			segments = append(segments, Segment{
				Speaker: "SPEAKER_" + u.Speaker,
				Start:   float64(u.Start) / 1000.0,
				End:     float64(u.End) / 1000.0,
				Text:    u.Text,
			})
		}
	} else {
		// Fallback: single segment with full transcript
		segments = append(segments, Segment{
			Speaker: "SPEAKER_A",
			Start:   0.0,
			End:     0.0,
			Text:    transcript,
		})
	}

	return TranscriptResponse{
		Transcript:     transcript,
		Segments:       segments,
		Language:       language,
		ProcessingTime: processingTime,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
