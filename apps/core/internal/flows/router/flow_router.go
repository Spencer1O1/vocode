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
	content = trimClassifierJSONResponse(content)
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
	res = DisambiguateClassifierResult(in, res)
	if err := res.Validate(); err != nil {
		return Result{}, err
	}
	return res, nil
}

// trimClassifierJSONResponse strips markdown code fences and leading prose so the model output
// parses as one JSON object. Anthropic (and some OpenAI models) often wrap JSON in ``` fences.
func trimClassifierJSONResponse(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "```") {
		rest := s
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
			rest = rest[nl+1:]
		} else {
			rest = ""
		}
		if end := strings.LastIndex(rest, "```"); end >= 0 {
			s = strings.TrimSpace(rest[:end])
		} else {
			s = strings.TrimSpace(rest)
		}
	}
	if i := strings.IndexByte(s, '{'); i > 0 {
		s = s[i:]
	} else if i < 0 {
		return strings.TrimSpace(s)
	}
	if j := strings.LastIndexByte(s, '}'); j >= 0 && j < len(s)-1 {
		s = s[:j+1]
	}
	return strings.TrimSpace(s)
}
