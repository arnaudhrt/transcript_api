package main

// --- API response types (match Python FastAPI output) ---

type TranscriptResponse struct {
	Transcript     string    `json:"transcript"`
	Segments       []Segment `json:"segments"`
	Language       string    `json:"language"`
	ProcessingTime float64   `json:"processing_time"`
}

type Segment struct {
	Speaker string  `json:"speaker"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Text    string  `json:"text"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}

// --- AssemblyAI API types ---

type UploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type TranscriptRequest struct {
	AudioURL          string `json:"audio_url"`
	SpeakerLabels     bool   `json:"speaker_labels"`
	SpeakersExpected  *int   `json:"speakers_expected,omitempty"`
	SpeechModel       string `json:"speech_model"`
}

type TranscriptStatus struct {
	ID           string      `json:"id"`
	Status       string      `json:"status"`
	Text         *string     `json:"text"`
	Utterances   []Utterance `json:"utterances"`
	LanguageCode *string     `json:"language_code"`
	Error        *string     `json:"error"`
}

type Utterance struct {
	Speaker string `json:"speaker"`
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
	Text    string `json:"text"`
}
