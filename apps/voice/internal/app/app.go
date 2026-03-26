package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"vocoding.net/vocode/v2/apps/voice/internal/mic"
	"vocoding.net/vocode/v2/apps/voice/internal/stt"
)

type App struct {
	in  io.Reader
	out io.Writer

	mu sync.Mutex

	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	segmentSeconds int64
}

func New(in io.Reader, out io.Writer) *App {
	segmentSeconds := int64(5)
	if v := strings.TrimSpace(os.Getenv("VOCODE_VOICE_SEGMENT_SECONDS")); v != "" {
		if parsed, err := time.ParseDuration(v + "s"); err == nil {
			segmentSeconds = int64(parsed.Seconds())
		}
	}

	return &App{
		in:             in,
		out:            out,
		segmentSeconds: segmentSeconds,
	}
}

type Request struct {
	Type string `json:"type"`
}

type Event struct {
	Type    string `json:"type"`
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Text    string `json:"text,omitempty"`
}

func (a *App) Run() error {
	if err := a.write(Event{
		Type:    "ready",
		Version: "0.1.0",
	}); err != nil {
		return err
	}

	scanner := bufio.NewScanner(a.in)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			if werr := a.write(Event{
				Type:    "error",
				Message: fmt.Sprintf("invalid request json: %v", err),
			}); werr != nil {
				return werr
			}
			continue
		}

		switch req.Type {
		case "start":
			if err := a.write(Event{
				Type:  "state",
				State: "starting",
			}); err != nil {
				return err
			}

			// Demo mode: if VOCODE_VOICE_DEMO_TRANSCRIPT is set, emit a transcript
			// event immediately (no audio required).
			if demo := strings.TrimSpace(os.Getenv("VOCODE_VOICE_DEMO_TRANSCRIPT")); demo != "" {
				if err := a.write(Event{Type: "transcript", Text: demo}); err != nil {
					return err
				}
				continue
			}

			if a.running {
				// Already running; treat as idempotent.
				continue
			}

			apiKey := strings.TrimSpace(os.Getenv("ELEVENLABS_API_KEY"))
			if apiKey == "" {
				if err := a.write(Event{Type: "error", Message: "ELEVENLABS_API_KEY is not set"}); err != nil {
					return err
				}
				continue
			}

			ctx, cancel := context.WithCancel(context.Background())
			rec, err := mic.Start(ctx, mic.StartParams{SampleRateHz: 16000, Channels: 1})
			if err != nil {
				cancel()
				if err := a.write(Event{Type: "error", Message: fmt.Sprintf("failed to start microphone recorder: %v", err)}); err != nil {
					return err
				}
				continue
			}

			a.running = true
			a.cancel = cancel
			a.wg.Add(1)
			go func() {
				defer a.wg.Done()
				a.transcribeLoop(ctx, apiKey, rec)
			}()

			if err := a.write(Event{Type: "state", State: "listening"}); err != nil {
				return err
			}
		case "stop":
			if a.running {
				a.running = false
				if a.cancel != nil {
					a.cancel()
					a.cancel = nil
				}
				a.wg.Wait()
			}

			if err := a.write(Event{
				Type:  "state",
				State: "stopped",
			}); err != nil {
				return err
			}
		case "shutdown":
			if a.running {
				a.running = false
				if a.cancel != nil {
					a.cancel()
					a.cancel = nil
				}
				a.wg.Wait()
			}

			if err := a.write(Event{
				Type:  "state",
				State: "shutdown",
			}); err != nil {
				return err
			}
			return nil
		default:
			if err := a.write(Event{
				Type:    "error",
				Message: fmt.Sprintf("unknown request type %q", req.Type),
			}); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func (a *App) write(evt Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	enc := json.NewEncoder(a.out)
	return enc.Encode(evt)
}

func (a *App) transcribeLoop(ctx context.Context, apiKey string, rec *mic.Recorder) {
	defer func() {
		_ = rec.Stop()
	}()

	// If demo transcript is configured, bypass STT calls to ElevenLabs to save credits.
	// We still emit at most one transcript event (the first time the loop runs).
	if demo := strings.TrimSpace(os.Getenv("VOCODE_VOICE_DEMO_TRANSCRIPT")); demo != "" {
		_ = a.write(Event{Type: "transcript", Text: demo})
		return
	}

	bytesPerSecond := int64(16000 * 1 * 2) // 16kHz * mono * int16
	targetBytes := bytesPerSecond * a.segmentSeconds
	if targetBytes <= 0 {
		targetBytes = bytesPerSecond * 5
	}

	buf := make([]byte, 32*1024)
	var segment []byte

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
				text, terr := stt.TranscribeElevenLabs(apiKey, "audio/wav", wav)
				if terr != nil {
					_ = a.write(Event{Type: "error", Message: fmt.Sprintf("elevenlabs stt failed: %v", terr)})
				} else if strings.TrimSpace(text) != "" {
					_ = a.write(Event{Type: "transcript", Text: text})
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
