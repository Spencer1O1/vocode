package agent

import (
	"fmt"
	"strings"
)

type ScopeKind string

const (
	ScopeCurrentFunction ScopeKind = "current_function"
	ScopeCurrentFile     ScopeKind = "current_file"
	ScopeNamedSymbol     ScopeKind = "named_symbol"
	ScopeClarify         ScopeKind = "clarify"
)

type ScopeIntentResult struct {
	ScopeKind    ScopeKind
	SymbolName   string
	ClarifyQuestion string
}

func (r ScopeIntentResult) Validate() error {
	switch r.ScopeKind {
	case ScopeCurrentFunction, ScopeCurrentFile:
		if strings.TrimSpace(r.SymbolName) != "" || strings.TrimSpace(r.ClarifyQuestion) != "" {
			return fmt.Errorf("scope intent: %s must not set symbolName/clarifyQuestion", r.ScopeKind)
		}
		return nil
	case ScopeNamedSymbol:
		if strings.TrimSpace(r.SymbolName) == "" {
			return fmt.Errorf("scope intent: named_symbol requires symbolName")
		}
		if strings.TrimSpace(r.ClarifyQuestion) != "" {
			return fmt.Errorf("scope intent: named_symbol must not set clarifyQuestion")
		}
		return nil
	case ScopeClarify:
		if strings.TrimSpace(r.ClarifyQuestion) == "" {
			return fmt.Errorf("scope intent: clarify requires clarifyQuestion")
		}
		if strings.TrimSpace(r.SymbolName) != "" {
			return fmt.Errorf("scope intent: clarify must not set symbolName")
		}
		return nil
	default:
		return fmt.Errorf("scope intent: unknown scopeKind %q", r.ScopeKind)
	}
}

