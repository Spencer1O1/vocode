package symbols

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type SymbolRef struct {
	Path string
	Line int
	Kind string
}

type Resolver interface {
	ResolveSymbol(workspaceRoot, symbolName, symbolKind, hintPath string) ([]SymbolRef, error)
}

type RipgrepResolver struct{}

func NewRipgrepResolver() *RipgrepResolver {
	return &RipgrepResolver{}
}

func (r *RipgrepResolver) ResolveSymbol(workspaceRoot, symbolName, symbolKind, hintPath string) ([]SymbolRef, error) {
	root := strings.TrimSpace(workspaceRoot)
	name := strings.TrimSpace(symbolName)
	if root == "" || name == "" {
		return nil, nil
	}
	kind := strings.ToLower(strings.TrimSpace(symbolKind))
	if kind == "" {
		kind = "function"
	}

	// Broad lexical search, narrowed by a lightweight parse of rg output.
	pattern := regexp.QuoteMeta(name) + `\s*\(`
	cmd := exec.Command("rg", "-n", "--no-heading", "--glob", "*.{go,ts,tsx,js,jsx}", pattern, root)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil && stdout.Len() == 0 {
		// No matches is not fatal; rg exits 1 in that case.
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	out := make([]SymbolRef, 0, len(lines))
	seen := map[string]bool{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 2 {
			continue
		}
		p := filepath.Clean(parts[0])
		key := p + ":" + parts[1]
		if seen[key] {
			continue
		}
		seen[key] = true
		lineNo := 0
		_, _ = fmt.Sscanf(parts[1], "%d", &lineNo)
		out = append(out, SymbolRef{Path: p, Line: lineNo, Kind: kind})
	}

	// If a hint path exists, bias to that file first.
	if hint := strings.TrimSpace(hintPath); hint != "" {
		hint = filepath.Clean(hint)
		biased := make([]SymbolRef, 0, len(out))
		for _, m := range out {
			if samePath(m.Path, hint) {
				biased = append(biased, m)
			}
		}
		if len(biased) > 0 {
			return biased, nil
		}
	}
	return out, nil
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}
