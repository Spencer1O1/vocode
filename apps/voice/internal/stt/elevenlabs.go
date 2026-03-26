package stt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

const elevenLabsSpeechToTextURL = "https://api.elevenlabs.io/v1/speech-to-text"

// TranscribeElevenLabs converts audio bytes into transcript text using ElevenLabs STT.
//
// The input bytes may be any format supported by ElevenLabs (e.g. webm/opus if
// coming from browser MediaRecorder). We pass the provided mimeType as the
// multipart file content type.
func TranscribeElevenLabs(apiKey string, modelID string, mimeType string, audio []byte, previousText string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ELEVENLABS_API_KEY is empty")
	}
	if modelID == "" {
		modelID = "scribe_v2"
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("audio is empty")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Match the JS SDK request field names.
	_ = writer.WriteField("model_id", modelID)
	_ = writer.WriteField("language_code", "eng")
	_ = writer.WriteField("tag_audio_events", "true")
	_ = writer.WriteField("diarize", "true")
	if previousText != "" {
		_ = writer.WriteField("previous_text", previousText)
	}

	contentType := mimeType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="file"; filename="audio"`},
		"Content-Type":        {contentType},
	})
	if err != nil {
		return "", err
	}
	if _, err := part.Write(audio); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 240 * time.Second}
	req, err := http.NewRequest("POST", elevenLabsSpeechToTextURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(resp.Body)
		return "", fmt.Errorf("elevenlabs stt http %d: %s", resp.StatusCode, b.String())
	}

	var parsed speechToTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	if parsed.Text != "" {
		return parsed.Text, nil
	}
	if len(parsed.Transcripts) > 0 {
		out := ""
		for i, t := range parsed.Transcripts {
			if i > 0 && out != "" {
				out += "\n"
			}
			out += t.Text
		}
		return out, nil
	}
	if parsed.Message != "" {
		return parsed.Message, nil
	}

	// Fallback: empty string indicates "no transcript".
	return "", nil
}

type speechToTextResponse struct {
	Text        string `json:"text"`
	Transcripts []struct {
		Text string `json:"text"`
	} `json:"transcripts"`
	Message string `json:"message"`
}

