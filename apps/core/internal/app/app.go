package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	"vocoding.net/vocode/v2/apps/core/internal/agent/anthropic"
	"vocoding.net/vocode/v2/apps/core/internal/agent/openai"
	"vocoding.net/vocode/v2/apps/core/internal/flows/router"
	"vocoding.net/vocode/v2/apps/core/internal/rpc"
	"vocoding.net/vocode/v2/apps/core/internal/transcript"
)

type App struct {
	logger *log.Logger
	server *rpc.Server
}

// New constructs the core daemon runtime.
func New(opts Options) (*App, error) {
	var stdin io.Reader = opts.Stdin
	var stdout io.Writer = opts.Stdout

	provider := strings.ToLower(strings.TrimSpace(os.Getenv("VOCODE_AGENT_PROVIDER")))
	var modelClient agent.ModelClient
	var err error
	switch provider {
	case "openai":
		modelClient, err = openai.NewFromEnv()
	case "anthropic":
		modelClient, err = anthropic.NewFromEnv()
	default:
		return nil, fmt.Errorf(
			`vocode agent: VOCODE_AGENT_PROVIDER must be "openai" or "anthropic" (got %q)`,
			os.Getenv("VOCODE_AGENT_PROVIDER"),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("vocode agent: %w", err)
	}
	flowRouter := router.NewFlowRouter(modelClient)
	if opts.Logger != nil {
		switch provider {
		case "openai":
			opts.Logger.Printf("vocode agent: using OpenAI model for flow routing and edits")
		case "anthropic":
			opts.Logger.Printf("vocode agent: using Anthropic model for flow routing and edits")
		}
	}
	voiceService := transcript.NewService(flowRouter, modelClient)

	r := rpc.NewRouter(opts.Logger)
	for _, def := range rpc.BuildHandlers(voiceService) {
		r.Register(def.Method, def.Handler)
	}

	server := rpc.NewServer(rpc.ServerOptions{
		Logger: opts.Logger,
		Stdin:  stdin,
		Stdout: stdout,
		Router: r,
	})

	voiceService.SetHostApplyClient(server)

	return &App{
		logger: opts.Logger,
		server: server,
	}, nil
}

func (a *App) Run() error {
	if a.logger != nil {
		a.logger.Println("vocode-cored starting...")
	}
	return a.server.Run()
}
