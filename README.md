# Meeting Transcription API (Go)

Go rewrite of the Python FastAPI speech-to-text service. Uploads audio to AssemblyAI for transcription with speaker diarization.

## Setup

```bash
cp .env.example .env
# Add your AssemblyAI API key to .env
go mod tidy
go run .
```

## Endpoints

- `GET /health` — Health check
- `POST /upload` — Upload audio file for transcription
  - Form fields: `file` (required), `speakers_expected` (optional int)
  - Allowed types: mp3, wav, m4a, mp4, avi, mov
  - Max size: 100MB

## Docker

```bash
docker compose up --build
```

## Response format

```json
{
  "transcript": "Full transcript text...",
  "segments": [
    {"speaker": "SPEAKER_A", "start": 0.0, "end": 5.2, "text": "..."}
  ],
  "language": "en",
  "processing_time": 45.2
}
```
