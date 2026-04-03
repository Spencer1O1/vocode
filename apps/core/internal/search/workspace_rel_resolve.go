package search

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolveWorkspaceRelativePath maps a slash-separated path relative to workspace root to an absolute
// path that exists on disk. Each segment is matched with: exact name, then case-insensitive unique,
// then [NormalizePathTokenForMatch] unique (handles Evade vs evade, my_utils vs "my utils" from the model).
// Returns ok false if the path cannot be resolved (caller may fall back to a literal join).
func ResolveWorkspaceRelativePath(root, rel string) (abs string, ok bool) {
	root = filepath.Clean(strings.TrimSpace(root))
	rel = strings.TrimSpace(rel)
	if root == "" {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return root, true
	}
	var parts []string
	for _, p := range strings.Split(rel, "/") {
		p = strings.TrimSpace(p)
		if p == "" || p == "." {
			continue
		}
		if p == ".." {
			return "", false
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return root, true
	}

	cur := root
	for i, part := range parts {
		last := i == len(parts)-1
		entries, err := os.ReadDir(cur)
		if err != nil {
			return "", false
		}
		name, picked := pickPathSegment(entries, part, last)
		if !picked {
			return "", false
		}
		cur = filepath.Join(cur, name)
	}
	return filepath.Clean(cur), true
}

func pickPathSegment(entries []fs.DirEntry, part string, isLast bool) (name string, ok bool) {
	var dirs, files []fs.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		} else if isLast {
			files = append(files, e)
		}
	}

	if n, ok := pickUniqueName(dirs, part, matchExact); ok {
		return n, true
	}
	if isLast {
		if n, ok := pickUniqueName(files, part, matchExact); ok {
			return n, true
		}
	}
	if n, ok := pickUniqueName(dirs, part, matchEqualFold); ok {
		return n, true
	}
	if n, ok := pickWithFoldDisambiguate(dirs, part); ok {
		return n, true
	}
	if isLast {
		if n, ok := pickUniqueName(files, part, matchEqualFold); ok {
			return n, true
		}
		if n, ok := pickWithFoldDisambiguate(files, part); ok {
			return n, true
		}
	}
	if n, ok := pickUniqueName(dirs, part, matchNorm); ok {
		return n, true
	}
	if isLast {
		if n, ok := pickUniqueName(files, part, matchNorm); ok {
			return n, true
		}
	}
	return "", false
}

func matchExact(name, part string) bool { return name == part }

func matchEqualFold(name, part string) bool { return strings.EqualFold(name, part) }

func matchNorm(name, part string) bool {
	return NormalizePathTokenForMatch(name) == NormalizePathTokenForMatch(part)
}

func pickUniqueName(entries []fs.DirEntry, part string, pred func(name, part string) bool) (string, bool) {
	var hits []string
	for _, e := range entries {
		if pred(e.Name(), part) {
			hits = append(hits, e.Name())
		}
	}
	if len(hits) != 1 {
		return "", false
	}
	return hits[0], true
}

// pickWithFoldDisambiguate handles multiple case variants (e.g. Evade and evade on Linux).
func pickWithFoldDisambiguate(entries []fs.DirEntry, part string) (string, bool) {
	var hits []string
	for _, e := range entries {
		if matchEqualFold(e.Name(), part) {
			hits = append(hits, e.Name())
		}
	}
	if len(hits) == 0 {
		return "", false
	}
	if len(hits) == 1 {
		return hits[0], true
	}
	if n, ok := disambiguateNamesByNorm(hits, part); ok {
		return n, true
	}
	sort.Strings(hits)
	return hits[0], true
}

func disambiguateNamesByNorm(names []string, part string) (string, bool) {
	want := NormalizePathTokenForMatch(part)
	var out []string
	for _, n := range names {
		if NormalizePathTokenForMatch(n) == want {
			out = append(out, n)
		}
	}
	if len(out) != 1 {
		return "", false
	}
	return out[0], true
}
