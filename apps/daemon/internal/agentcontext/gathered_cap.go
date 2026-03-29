package agentcontext

import (
	"path/filepath"
	"strings"
)

// ApplyGatheredRollingCap trims excerpts, notes, and symbols so context stays bounded.
// The excerpt whose path matches activeFile (after filepath.Clean) is never removed entirely;
// if byte pressure remains with only that excerpt, its content is truncated by runes.
func ApplyGatheredRollingCap(g Gathered, activeFile string, maxBytes int, maxExcerpts int) Gathered {
	if maxExcerpts <= 0 {
		maxExcerpts = 12
	}
	if maxBytes <= 0 {
		maxBytes = 120_000
	}
	active := filepath.Clean(strings.TrimSpace(activeFile))

	var activeEx *FileExcerpt
	var others []FileExcerpt
	for _, ex := range g.Excerpts {
		p := filepath.Clean(strings.TrimSpace(ex.Path))
		if active != "" && p == active && activeEx == nil {
			c := ex
			activeEx = &c
			continue
		}
		others = append(others, ex)
	}

	for {
		g.Excerpts = joinExcerpts(activeEx, others)
		exN := len(g.Excerpts)
		if exN <= maxExcerpts && EstimatedGatheredBytes(g) <= maxBytes {
			break
		}
		if len(others) == 0 {
			break
		}
		others = others[1:]
	}

	g.Excerpts = joinExcerpts(activeEx, others)
	const maxNotes = 8
	if len(g.Notes) > maxNotes {
		g.Notes = g.Notes[len(g.Notes)-maxNotes:]
	}
	const maxSymbols = 40
	if len(g.Symbols) > maxSymbols {
		g.Symbols = g.Symbols[len(g.Symbols)-maxSymbols:]
	}

	for EstimatedGatheredBytes(g) > maxBytes {
		if len(g.Excerpts) == 0 {
			break
		}
		first := filepath.Clean(strings.TrimSpace(g.Excerpts[0].Path))
		if active != "" && first == active {
			runes := []rune(g.Excerpts[0].Content)
			if len(runes) <= 256 {
				break
			}
			g.Excerpts[0].Content = string(runes[:len(runes)*2/3])
			continue
		}
		g.Excerpts = g.Excerpts[1:]
	}
	return g
}

func joinExcerpts(active *FileExcerpt, others []FileExcerpt) []FileExcerpt {
	if active == nil {
		out := make([]FileExcerpt, len(others))
		copy(out, others)
		return out
	}
	out := make([]FileExcerpt, 0, 1+len(others))
	out = append(out, *active)
	out = append(out, others...)
	return out
}
