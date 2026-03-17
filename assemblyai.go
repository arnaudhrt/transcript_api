package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.assemblyai.com"

type AssemblyAIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAssemblyAIClient(apiKey string) *AssemblyAIClient {
	return &AssemblyAIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// Transcribe uploads an audio file and returns the completed transcript.
func (c *AssemblyAIClient) Transcribe(ctx context.Context, filePath string, speakersExpected *int) (*TranscriptStatus, error) {
	// 1. Upload file
	slog.Info("uploading file to AssemblyAI", "path", filePath)
	uploadURL, err := c.uploadFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	slog.Info("file uploaded", "upload_url", uploadURL)

	// 2. Submit transcription
	slog.Info("submitting transcription", "speakers_expected", speakersExpected)
	transcriptID, err := c.submitTranscription(ctx, uploadURL, speakersExpected)
	if err != nil {
		return nil, fmt.Errorf("submit failed: %w", err)
	}
	slog.Info("transcription submitted", "id", transcriptID)

	// 3. Poll for completion
	return c.pollTranscription(ctx, transcriptID)
}

func (c *AssemblyAIClient) uploadFile(ctx context.Context, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v2/upload", f)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload returned %d: %s", resp.StatusCode, string(body))
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.UploadURL, nil
}

func (c *AssemblyAIClient) submitTranscription(ctx context.Context, audioURL string, speakersExpected *int) (string, error) {
	body := TranscriptRequest{
		AudioURL:         audioURL,
		SpeakerLabels:    true,
		SpeakersExpected: speakersExpected,
		SpeechModel:      "best",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v2/transcript", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("submit returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result TranscriptStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (c *AssemblyAIClient) pollTranscription(ctx context.Context, id string) (*TranscriptStatus, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v2/transcript/"+id, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", c.apiKey)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, err
			}

			var result TranscriptStatus
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()

			slog.Info("poll status", "id", id, "status", result.Status)

			switch result.Status {
			case "completed":
				return &result, nil
			case "error":
				errMsg := "unknown error"
				if result.Error != nil {
					errMsg = *result.Error
				}
				return nil, fmt.Errorf("transcription failed: %s", errMsg)
			}
			// "queued" or "processing" — keep polling
		}
	}
}
