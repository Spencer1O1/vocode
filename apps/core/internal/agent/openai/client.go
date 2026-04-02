// Package openai implements agent.ModelClient via OpenAI Chat Completions (JSON schema response_format).
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	aiprompt "vocoding.net/vocode/v2/apps/core/internal/agent/prompt"
)

type Client struct {
	HTTPClient *http.Client
	APIKey     string
	BaseURL    string
	Model      string
}

func NewFromEnv() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}
	base := strings.TrimSpace(os.Getenv("VOCODE_OPENAI_BASE_URL"))
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(os.Getenv("VOCODE_OPENAI_MODEL"))
	if model == "" {
		return nil, fmt.Errorf("VOCODE_OPENAI_MODEL is not set")
	}
	return &Client{
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		APIKey:     key,
		BaseURL:    base,
		Model:      model,
	}, nil
}

func (c *Client) ClassifyFlow(ctx context.Context, in agent.ClassifierContext) (agent.ClassifierResult, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return agent.ClassifierResult{}, fmt.Errorf("openai: missing API key")
	}
	userBytes, err := aiprompt.FlowClassifierUserJSON(in)
	if err != nil {
		return agent.ClassifierResult{}, fmt.Errorf("openai: prompt: %w", err)
	}
	temp := 0.0
	body := chatCompletionsRequest{
		Model:       c.Model,
		Temperature: &temp,
		Messages: []chatMessage{
			{Role: "system", Content: aiprompt.FlowClassifierSystem(in.Flow)},
			{Role: "user", Content: string(userBytes)},
		},
		ResponseFormat: chatResponseFormatFlowClassifier(in.Flow),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return agent.ClassifierResult{}, err
	}
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return agent.ClassifierResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return agent.ClassifierResult{}, fmt.Errorf("openai: request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return agent.ClassifierResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return agent.ClassifierResult{}, fmt.Errorf("openai: HTTP %s: %s", resp.Status, truncateForErr(respBody, 512))
	}
	var parsed chatCompletionsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return agent.ClassifierResult{}, fmt.Errorf("openai: decode response: %w", err)
	}
	if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return agent.ClassifierResult{}, fmt.Errorf("openai: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return agent.ClassifierResult{}, fmt.Errorf("openai: empty choices")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return agent.ClassifierResult{}, fmt.Errorf("openai: empty message content")
	}
	var raw struct {
		Route string `json:"route"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return agent.ClassifierResult{}, fmt.Errorf("openai: decode classifier: %w", err)
	}
	res := agent.ClassifierResult{
		Flow:  in.Flow,
		Route: strings.TrimSpace(raw.Route),
	}
	if err := res.Validate(); err != nil {
		return agent.ClassifierResult{}, err
	}
	return res, nil
}

type chatCompletionsRequest struct {
	Model          string          `json:"model"`
	Temperature    *float64        `json:"temperature,omitempty"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func truncateForErr(b []byte, max int) string {
	s := strings.TrimSpace(string(b))
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
