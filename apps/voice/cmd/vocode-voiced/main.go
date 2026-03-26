package main

import (
	"log"
	"os"

	"vocoding.net/vocode/v2/apps/voice/internal/app"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	a := app.New(os.Stdin, os.Stdout)
	if err := a.Run(); err != nil {
		logger.Fatalf("voice sidecar exited with error: %v", err)
	}
}
