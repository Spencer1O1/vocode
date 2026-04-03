### Agent config refactor (provider/model/base URL)

This document describes a future refactor to stop using environment variables for the agent provider/model/base URL selection (except for API keys), and instead pass an explicit configuration from the extension to `vocode-cored`, similar to how `daemonConfig` works for transcript caps (RPC field name is historical).

#### 1. Introduce explicit agent config in the core

- Add a new type in the core agent package (e.g. `apps/core/internal/agent/config.go`):

```go
type AgentProvider string

const (
	AgentProviderOpenAI    AgentProvider = "openai"
	AgentProviderAnthropic AgentProvider = "anthropic"
)

type AgentConfig struct {
	Provider         AgentProvider
	OpenAIModel      string
	OpenAIBaseURL    string
	AnthropicModel   string
	AnthropicBaseURL string
}
```

- Add a constructor that takes this config instead of reading env:

```go
// NewWithConfig constructs an Agent from an explicit config.
func NewWithConfig(cfg AgentConfig, logger *log.Logger) (*Agent, error) {
	switch cfg.Provider {
	case AgentProviderOpenAI:
		c, err := openai.NewWithConfig(openai.Config{
			APIKey:  os.Getenv("OPENAI_API_KEY"), // API key still from env
			Model:   cfg.OpenAIModel,
			BaseURL: cfg.OpenAIBaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("openai: %w", err)
		}
		return &Agent{model: c}, nil
	case AgentProviderAnthropic:
		c, err := anthropic.NewWithConfig(anthropic.Config{
			APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
			Model:   cfg.AnthropicModel,
			BaseURL: cfg.AnthropicBaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("anthropic: %w", err)
		}
		return &Agent{model: c}, nil
	default:
		return nil, fmt.Errorf("unknown agent provider %q", cfg.Provider)
	}
}
```

- Keep the existing `New` / `NewFromEnv` paths as thin shims for CLI and legacy use, but have them build an `AgentConfig` from env and call `NewWithConfig`.

#### 2. Add config-aware constructors in model clients

- In `apps/core/internal/agent/openai/client.go` introduce:

```go
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

func NewWithConfig(cfg Config) (*Client, error) {
	// validate + default BaseURL / Model without reading env
}

func NewFromEnv() (*Client, error) {
	return NewWithConfig(Config{
		APIKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		BaseURL: strings.TrimSpace(os.Getenv("VOCODE_OPENAI_BASE_URL")),
		Model:   strings.TrimSpace(os.Getenv("VOCODE_OPENAI_MODEL")),
	})
}
```

- Mirror the same pattern for the Anthropic client.

#### 3. Change core bootstrap to accept explicit config

- In `apps/core/internal/app/app.go`:
  - Replace direct env reads with a path that can accept an `AgentConfig`.
  - For now, build `AgentConfig` from env inside bootstrap and delegate to `agent.NewWithConfig`, so behaviour stays identical while the extension is still passing provider/model/base URL via env.
  - Later, add a way to pass `AgentConfig` into `New` without env (e.g. command-line flags or a small JSON config file).

#### 4. Update the extension to stop using env for provider/model/base URL

- In `apps/vscode-extension/src/config/spawn-env.ts`:
  - Remove the `CONFIG_TO_ENV` bindings for:
    - `daemonAgentProvider` → `VOCODE_AGENT_PROVIDER`
    - `daemonOpenaiModel` → `VOCODE_OPENAI_MODEL`
    - `daemonOpenaiBaseUrl` → `VOCODE_OPENAI_BASE_URL`
    - `daemonAnthropicModel` → `VOCODE_ANTHROPIC_MODEL`
    - `daemonAnthropicBaseUrl` → `VOCODE_ANTHROPIC_BASE_URL`
  - Keep API keys in env (`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `ELEVENLABS_API_KEY`) exactly as today.

- When spawning `vocode-cored`, pass a small JSON config blob or command-line flags that mirror `AgentConfig`:
  - Values come from `vscode.workspace.getConfiguration("vocode")`:
    - `daemonAgentProvider`
    - `daemonOpenaiModel`
    - `daemonOpenaiBaseUrl`
    - `daemonAnthropicModel`
    - `daemonAnthropicBaseUrl`

#### 5. Tests and validation

- Add tests around:
  - `openai.NewWithConfig` and `anthropic.NewWithConfig` (no env reads).
  - `agent.NewWithConfig`: selects OpenAI or Anthropic based on `Provider`; returns an error when the client constructor fails or the provider is unknown.
  - End-to-end executor tests that construct an `Agent` via `NewWithConfig` and run a transcript without relying on provider/model/base URL env vars.

- Once the extension is updated and tests are green:
  - Consider marking env-based provider/model/base URL selection as deprecated in comments.
  - Optionally remove those env reads from the main core path, keeping them only for direct `go run`/CLI workflows if needed.
