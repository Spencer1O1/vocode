package agentcontext

import (
	"path/filepath"
	"strings"
)

// UpsertGatheredExcerpt replaces an excerpt for path or appends it. Path is cleaned for comparison.
func UpsertGatheredExcerpt(g Gathered, absPath, content string) Gathered {
	p := filepath.Clean(strings.TrimSpace(absPath))
	if p == "" {
		return g
	}
	for i := range g.Excerpts {
		if filepath.Clean(g.Excerpts[i].Path) == p {
			g.Excerpts[i].Path = p
			g.Excerpts[i].Content = content
			return g
		}
	}
	g.Excerpts = append(g.Excerpts, FileExcerpt{Path: p, Content: content})
	return g
}

// SeedGatheredActiveFile ensures the active file excerpt is present (bootstrap for turn 0).
func SeedGatheredActiveFile(g Gathered, absPath string) Gathered {
	ex := ReadActiveFileExcerpt(absPath)
	if ex == "" {
		return g
	}
	return UpsertGatheredExcerpt(g, absPath, ex)
}
