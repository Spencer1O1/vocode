package main

import (
	"log"
	"os"

	"vocoding.net/vocode/v2/apps/daemon/internal/app"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	a, err := app.New(app.Options{
		Logger: logger,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		logger.Fatalf("failed to create app: %v", err)
	}

	if err := a.Run(); err != nil {
		logger.Fatalf("daemon exited with error: %v", err)
	}
}
