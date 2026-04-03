package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
)

// FlowRouter maps a transcript to a route for the active flow.
type FlowRouter struct {
	Model agent.ModelClient
}

// NewFlowRouter returns a router backed by a cloud model client. Model must be non-nil.
func NewFlowRouter(model agent.ModelClient) *FlowRouter {
	return &FlowRouter{Model: model}
}

// ClassifyFlow returns route classification and structured fields (e.g. search_query for rg).
func (r *FlowRouter) ClassifyFlow(ctx context.Context, in Context) (Result, error) {
	if r == nil {
		return Result{}, fmt.Errorf("router: nil FlowRouter")
	}
	if r.Model == nil {
		return Result{}, fmt.Errorf("router: AI model is not configured")
	}
	return classifyWithModel(ctx, r.Model, in)
}

func classifyWithModel(ctx context.Context, m agent.ModelClient, in Context) (Result, error) {
	userBytes, err := ClassifierUserJSON(in)
	if err != nil {
		return Result{}, err
	}
	schema := ClassifierResponseJSONSchema(in.Flow)
	content, err := m.Call(ctx, agent.CompletionRequest{
		System:     ClassifierSystem(in.Flow),
		User:       string(userBytes),
		JSONSchema: schema,
	})
	if err != nil {
		return Result{}, err
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return Result{}, ErrEmptyModelContent
	}
	var raw struct {
		Route            string `json:"route"`
		SearchQuery      string `json:"search_query"`
		SearchSymbolKind string `json:"search_symbol_kind"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return Result{}, fmt.Errorf("router: decode classifier json: %w", err)
	}
	res := Result{
		Flow:             in.Flow,
		Route:            strings.TrimSpace(raw.Route),
		SearchQuery:      strings.TrimSpace(raw.SearchQuery),
		SearchSymbolKind: strings.TrimSpace(strings.ToLower(raw.SearchSymbolKind)),
	}
	if err := res.Validate(); err != nil {
		return Result{}, err
	}
	return res, nil
}
