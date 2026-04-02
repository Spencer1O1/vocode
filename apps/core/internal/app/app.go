package app

import (
	"io"
	"log"
	"os"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	"vocoding.net/vocode/v2/apps/core/internal/agent/openai"
	"vocoding.net/vocode/v2/apps/core/internal/agent/stub"
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

	agentRuntime := agent.New(selectModelClient(opts.Logger))
	voiceService := transcript.NewService(agentRuntime)

	router := rpc.NewRouter(opts.Logger)
	for _, def := range rpc.BuildHandlers(voiceService) {
		router.Register(def.Method, def.Handler)
	}

	server := rpc.NewServer(rpc.ServerOptions{
		Logger: opts.Logger,
		Stdin:  stdin,
		Stdout: stdout,
		Router: router,
	})

	voiceService.SetHostApplyClient(server)

	return &App{
		logger: opts.Logger,
		server: server,
	}, nil
}

func selectModelClient(logger *log.Logger) agent.ModelClient {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("VOCODE_AGENT_PROVIDER")))
	switch provider {
	case "", "stub":
		return stub.New()
	case "openai":
		c, err := openai.NewFromEnv()
		if err != nil {
			if logger != nil {
				logger.Printf("vocode agent: OpenAI unavailable (%v); using stub model client", err)
			}
			return stub.New()
		}
		if logger != nil {
			logger.Printf("vocode agent: using OpenAI model client")
		}
		return c
	default:
		if logger != nil {
			logger.Printf("vocode agent: unknown VOCODE_AGENT_PROVIDER %q; using stub model client", provider)
		}
		return stub.New()
	}
}

func (a *App) Run() error {
	if a.logger != nil {
		a.logger.Println("vocode-cored starting...")
	}
	return a.server.Run()
}

