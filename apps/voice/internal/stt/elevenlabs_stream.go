package stt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const elevenLabsSpeechToTextStreamURL = "wss://api.elevenlabs.io/v1/speech-to-text/stream"

type StreamingEvent struct {
	Text  string
	Error error
}

type ElevenLabsStreamingClient struct {
	conn   *websocket.Conn
	events chan StreamingEvent
	done   chan struct{}
	mu     sync.Mutex
}

func NewElevenLabsStreamingClient(ctx context.Context, apiKey string, sampleRate int) (*ElevenLabsStreamingClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("ELEVENLABS_API_KEY is empty")
	}
	if sampleRate <= 0 {
		sampleRate = 16000
	}

	wsURL, err := url.Parse(elevenLabsSpeechToTextStreamURL)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Set("xi-api-key", apiKey)

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), header)
	if err != nil {
		return nil, err
	}

	c := &ElevenLabsStreamingClient{
		conn:   conn,
		events: make(chan StreamingEvent, 16),
		done:   make(chan struct{}),
	}

	go c.readLoop()
	return c, nil
}

func (c *ElevenLabsStreamingClient) Events() <-chan StreamingEvent {
	return c.events
}

func (c *ElevenLabsStreamingClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.done:
		return nil
	default:
		close(c.done)
	}

	err := c.conn.Close()
	return err
}

func (c *ElevenLabsStreamingClient) SendInputAudioChunk(pcm []byte, commit bool, previousText string) error {
	if len(pcm) == 0 {
		return nil
	}

	msg := map[string]any{
		"message_type": "input_audio_chunk",
		"audio_base_64": base64.StdEncoding.EncodeToString(pcm),
		"sample_rate":  16000,
		"commit":       commit,
	}
	if strings.TrimSpace(previousText) != "" {
		msg["previous_text"] = previousText
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

func (c *ElevenLabsStreamingClient) readLoop() {
	defer close(c.events)

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			select {
			case c.events <- StreamingEvent{Error: err}:
			default:
			}
			return
		}

		text := extractStreamingText(data)
		if strings.TrimSpace(text) == "" {
			continue
		}
		select {
		case c.events <- StreamingEvent{Text: text}:
		default:
		}
	}
}

func extractStreamingText(data []byte) string {
	var generic map[string]any
	if err := json.Unmarshal(data, &generic); err != nil {
		return ""
	}

	if v, ok := generic["text"].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	if tr, ok := generic["transcript"].(string); ok && strings.TrimSpace(tr) != "" {
		return tr
	}
	return ""
}
