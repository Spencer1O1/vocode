package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	"vocoding.net/vocode/v2/apps/voice/internal/mic"
	"vocoding.net/vocode/v2/apps/voice/internal/stt"
)

func (a *App) transcribeLoop(ctx context.Context, apiKey string, modelID string, rec *mic.Recorder) {
	defer func() {
		_ = rec.Stop()
	}()

	if sttMode() == "stream" {
		a.transcribeLoopStream(ctx, apiKey, modelID, rec)
		return
	}
	a.transcribeLoopBatch(ctx, apiKey, modelID, rec)
}

func (a *App) transcribeLoopBatch(ctx context.Context, apiKey string, modelID string, rec *mic.Recorder) {
	bytesPerSecond := int64(16000 * 1 * 2) // 16kHz * mono * int16
	targetBytes := bytesPerSecond * a.segmentSeconds
	if targetBytes <= 0 {
		targetBytes = bytesPerSecond * 5
	}

	buf := make([]byte, 32*1024)
	var segment []byte
	contextWindow := newUtteranceWindow(4, 1200)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := rec.PCMReader().Read(buf)
		if n > 0 {
			segment = append(segment, buf[:n]...)
		}

		if int64(len(segment)) >= targetBytes {
			wav, werr := mic.EncodeWavPCM16LE(segment, 16000, 1)
			if werr != nil {
				_ = a.write(Event{Type: "error", Message: fmt.Sprintf("failed to encode wav: %v", werr)})
			} else {
				text, terr := stt.TranscribeElevenLabs(apiKey, modelID, "audio/wav", wav, contextWindow.PreviousText())
				if terr != nil {
					_ = a.write(Event{Type: "error", Message: fmt.Sprintf("elevenlabs stt failed: %v", terr)})
				} else if strings.TrimSpace(text) != "" {
					_ = a.write(Event{Type: "transcript", Text: text})
					contextWindow.AddUtterance(text)
				}
			}
			segment = nil
		}

		if err != nil {
			if err == io.EOF {
				return
			}
			_ = a.write(Event{Type: "error", Message: fmt.Sprintf("microphone read failed: %v", err)})
			return
		}
	}
}

func (a *App) transcribeLoopStream(ctx context.Context, apiKey string, modelID string, rec *mic.Recorder) {
	client, err := stt.NewElevenLabsStreamingClient(ctx, apiKey, modelID, 16000)
	if err != nil {
		_ = a.write(Event{Type: "error", Message: fmt.Sprintf("failed to start elevenlabs streaming stt: %v", err)})
		return
	}
	defer func() {
		_ = client.Close()
	}()

	bytesPerSecond := int64(16000 * 1 * 2) // 16kHz * mono * int16
	minChunkBytes := bytesPerSecond * int64(streamMinChunkMS()) / 1000
	if minChunkBytes <= 0 {
		minChunkBytes = 6400
	}
	maxChunkBytes := bytesPerSecond * int64(streamMaxChunkMS()) / 1000
	if maxChunkBytes < minChunkBytes {
		maxChunkBytes = minChunkBytes
	}

	buf := make([]byte, 32*1024)
	var chunk []byte
	contextWindow := newUtteranceWindow(4, 1200)
	vad := newLocalVAD(16000, int(minChunkBytes), int(maxChunkBytes), streamMaxUtteranceMS())

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, readErr := rec.PCMReader().Read(buf)
		if n > 0 {
			chunk = append(chunk, buf[:n]...)
		}

		for len(chunk) >= vad.frameBytes {
			frame := chunk[:vad.frameBytes]
			chunk = chunk[vad.frameBytes:]
			for _, c := range vad.process(frame) {
				if err := client.SendInputAudioChunk(c.pcm, c.commit, contextWindow.PreviousText()); err != nil {
					_ = a.write(Event{Type: "error", Message: fmt.Sprintf("elevenlabs streaming send failed: %v", err)})
					return
				}
			}
		}

		select {
		case evt, ok := <-client.Events():
			if !ok {
				return
			}
			if evt.Error != nil {
				_ = a.write(Event{Type: "error", Message: fmt.Sprintf("elevenlabs streaming stt failed: %v", evt.Error)})
				return
			}
			if strings.TrimSpace(evt.Text) != "" {
				_ = a.write(Event{Type: "transcript", Text: evt.Text})
				if evt.IsFinal {
					contextWindow.AddUtterance(evt.Text)
				}
			}
		default:
		}

		if readErr != nil {
			if readErr == io.EOF {
				for _, c := range vad.flush() {
					if err := client.SendInputAudioChunk(c.pcm, c.commit, contextWindow.PreviousText()); err != nil {
						_ = a.write(Event{Type: "error", Message: fmt.Sprintf("elevenlabs streaming send failed: %v", err)})
						return
					}
				}
				return
			}
			_ = a.write(Event{Type: "error", Message: fmt.Sprintf("microphone read failed: %v", readErr)})
			return
		}
	}
}
