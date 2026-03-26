package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type App struct {
	in  io.Reader
	out io.Writer
}

func New(in io.Reader, out io.Writer) *App {
	return &App{in: in, out: out}
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

			// Demo mode: if VOCOIDE_VOICE_DEMO_TRANSCRIPT is set, emit a transcript
			// event. This helps validate the extension<->sidecar wiring before
			// microphone/STT is implemented.
			if demo := strings.TrimSpace(os.Getenv("VOCODE_VOICE_DEMO_TRANSCRIPT")); demo != "" {
				if err := a.write(Event{
					Type: "transcript",
					Text: demo,
				}); err != nil {
					return err
				}
			} else {
				if err := a.write(Event{
					Type:    "error",
					Message: "microphone capture is not implemented yet in apps/voice",
				}); err != nil {
					return err
				}
			}
		case "stop":
			if err := a.write(Event{
				Type:  "state",
				State: "stopped",
			}); err != nil {
				return err
			}
		case "shutdown":
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
	enc := json.NewEncoder(a.out)
	return enc.Encode(evt)
}

