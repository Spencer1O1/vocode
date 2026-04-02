// Package stub provides a fixed-response agent.ModelClient for dev wiring.
package stub

import (
	"context"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

type Client struct{}

func New() *Client { return &Client{} }

func (*Client) ClassifyFlow(ctx context.Context, in agent.ClassifierContext) (agent.ClassifierResult, error) {
	_ = ctx
	t := strings.TrimSpace(strings.ToLower(in.Instruction))
	switch in.Flow {
	case flows.Select:
		return stubSelect(t), nil
	case flows.SelectFile:
		return stubSelectFile(t), nil
	default:
		return stubRoot(t), nil
	}
}

func stubRoot(t string) agent.ClassifierResult {
	if t == "" {
		return agent.ClassifierResult{Flow: flows.Root, Route: "irrelevant"}
	}
	if strings.HasPrefix(t, "find file ") || strings.HasPrefix(t, "find files ") ||
		strings.HasPrefix(t, "open file ") || strings.HasPrefix(t, "show file ") ||
		strings.HasPrefix(t, "file named ") || strings.HasPrefix(t, "locate file ") {
		return agent.ClassifierResult{Flow: flows.Root, Route: "select_file"}
	}
	if strings.HasPrefix(t, "find ") || strings.HasPrefix(t, "search ") || strings.HasPrefix(t, "where is ") || strings.HasPrefix(t, "locate ") {
		return agent.ClassifierResult{Flow: flows.Root, Route: "select"}
	}
	if strings.HasSuffix(t, "?") || strings.HasPrefix(t, "what ") || strings.HasPrefix(t, "why ") || strings.HasPrefix(t, "how ") {
		return agent.ClassifierResult{Flow: flows.Root, Route: "question"}
	}
	if globalExitLike(t) {
		return agent.ClassifierResult{Flow: flows.Root, Route: "control"}
	}
	return agent.ClassifierResult{Flow: flows.Root, Route: "irrelevant"}
}

func stubSelect(t string) agent.ClassifierResult {
	if t == "" {
		return agent.ClassifierResult{Flow: flows.Select, Route: "irrelevant"}
	}
	if globalExitLike(t) {
		return agent.ClassifierResult{Flow: flows.Select, Route: "control"}
	}
	if strings.Contains(t, "find file ") || strings.Contains(t, "open file ") || strings.Contains(t, "show file ") {
		return agent.ClassifierResult{Flow: flows.Select, Route: "select_file"}
	}
	if strings.Contains(t, "find ") || strings.Contains(t, "search ") {
		return agent.ClassifierResult{Flow: flows.Select, Route: "select"}
	}
	if strings.Contains(t, "next") || strings.Contains(t, "forward") ||
		strings.Contains(t, "back") || strings.Contains(t, "prev") {
		return agent.ClassifierResult{Flow: flows.Select, Route: "select_control"}
	}
	if strings.Contains(t, "delete") || strings.Contains(t, "remove") {
		return agent.ClassifierResult{Flow: flows.Select, Route: "delete"}
	}
	for _, w := range []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		if strings.Contains(t, w) {
			return agent.ClassifierResult{Flow: flows.Select, Route: "select_control"}
		}
	}
	if strings.Contains(t, "edit") || strings.Contains(t, "change") {
		return agent.ClassifierResult{Flow: flows.Select, Route: "edit"}
	}
	return agent.ClassifierResult{Flow: flows.Select, Route: "irrelevant"}
}

func stubSelectFile(t string) agent.ClassifierResult {
	if t == "" {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "irrelevant"}
	}
	if globalExitLike(t) {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "control"}
	}
	if strings.Contains(t, "next") || strings.Contains(t, "forward") ||
		strings.Contains(t, "back") || strings.Contains(t, "prev") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "select_file_control"}
	}
	for _, w := range []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		if strings.Contains(t, w) {
			return agent.ClassifierResult{Flow: flows.SelectFile, Route: "select_file_control"}
		}
	}
	if strings.Contains(t, "delete") || strings.Contains(t, "remove") || strings.Contains(t, "trash") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "delete"}
	}
	if strings.Contains(t, "open") || strings.Contains(t, "show") || strings.Contains(t, "reveal") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "open"}
	}
	if strings.Contains(t, "rename") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "rename"}
	}
	if strings.Contains(t, "move") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "move"}
	}
	if strings.Contains(t, "create") || strings.Contains(t, "new file") || strings.Contains(t, "new folder") {
		return agent.ClassifierResult{Flow: flows.SelectFile, Route: "create"}
	}
	return agent.ClassifierResult{Flow: flows.SelectFile, Route: "irrelevant"}
}

func globalExitLike(t string) bool {
	t = strings.TrimSpace(strings.ToLower(t))
	for _, w := range []string{"cancel", "exit", "close", "stop", "done", "quit", "leave", "abort"} {
		if strings.Contains(t, w) {
			return true
		}
	}
	return false
}
