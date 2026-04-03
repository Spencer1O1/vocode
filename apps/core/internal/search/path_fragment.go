package search

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var pathSearchSkipDirNames = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	".pnpm-store":  {},
	"vendor":       {},
	"dist":         {},
	"bin":          {},
	".turbo":       {},
	"__pycache__":  {},
	".venv":        {},
	"target":       {}, // Rust
	".idea":        {},
}

// PathMatch is one file or directory path from path-fragment discovery.
type PathMatch struct {
	Path  string
	IsDir bool
}

func pathFragmentMatches(rel, baseName, lowerFragment string) bool {
	return pathContainsFold(rel, lowerFragment) || pathContainsFold(baseName, lowerFragment)
}

// PathFragmentMatches lists files and directories under root whose relative path or base name contains
// fragment (case-insensitive). For select_file the fragment is a single basename. Not content search.
// Per-segment path resolution (e.g. voice move targets) uses [ResolveWorkspaceRelativePath] instead.
// Paths strictly inside a matched directory are omitted so listing a folder (e.g. "Res") does not
// also return every file under it. Prepending the workspace root (when the query names that folder)
// happens before pruning so children under that root are still dropped. maxPaths <= 0 defaults to 20.
func PathFragmentMatches(root, fragment string, maxPaths int) ([]PathMatch, error) {
	root = filepath.Clean(strings.TrimSpace(root))
	fragment = strings.TrimSpace(fragment)
	if root == "" || fragment == "" {
		return nil, nil
	}
	if maxPaths <= 0 {
		maxPaths = 20
	}
	lower := strings.ToLower(fragment)

	var matches []PathMatch
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != root {
				if _, skip := pathSearchSkipDirNames[name]; skip {
					return filepath.SkipDir
				}
			}
			if path == root {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if pathFragmentMatches(rel, name, lower) {
				matches = append(matches, PathMatch{Path: filepath.Clean(path), IsDir: true})
			}
			if len(matches) >= maxPaths*4 {
				return filepath.SkipAll
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		base := filepath.Base(path)
		if !pathFragmentMatches(rel, base, lower) {
			return nil
		}
		matches = append(matches, PathMatch{Path: filepath.Clean(path), IsDir: false})
		if len(matches) >= maxPaths*4 {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	matches = prependWorkspaceRootIfBasenameMatches(root, fragment, matches)
	matches = prunePathMatchesUnderMatchedDirs(matches)
	rankPathFragmentMatches(matches, fragment)
	return uniqueSortedPathMatchesCap(matches, maxPaths), nil
}

// isStrictDescendant reports whether child is inside parent (not equal).
func isStrictDescendant(child, parent string) bool {
	child = filepath.Clean(child)
	parent = filepath.Clean(parent)
	if child == parent || parent == "" || child == "" {
		return false
	}
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// prunePathMatchesUnderMatchedDirs removes files and nested dirs whose path lies under a matched
// directory. Otherwise a fragment like "Res" lists the folder plus every file under it because
// relative paths contain "Res" as a substring.
func prunePathMatchesUnderMatchedDirs(matches []PathMatch) []PathMatch {
	var dirPaths []string
	for _, m := range matches {
		if m.IsDir && m.Path != "" {
			dirPaths = append(dirPaths, filepath.Clean(m.Path))
		}
	}
	if len(dirPaths) == 0 {
		return matches
	}
	out := make([]PathMatch, 0, len(matches))
	for _, m := range matches {
		p := filepath.Clean(m.Path)
		if p == "" {
			continue
		}
		drop := false
		for _, d := range dirPaths {
			if p == d {
				break
			}
			if isStrictDescendant(p, d) {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, m)
		}
	}
	return out
}

// fileSearchQueryStopWords strips spoken noise from select_file search queries when matching basenames.
var fileSearchQueryStopWords = map[string]struct{}{
	"find": {}, "the": {}, "a": {}, "an": {}, "open": {}, "show": {}, "please": {},
	"goto": {}, "to": {}, "folder": {}, "folders": {}, "directory": {}, "directories": {},
	"file": {}, "path": {}, "need": {}, "want": {}, "where": {}, "is": {}, "my": {},
	"this": {}, "that": {}, "me": {}, "can": {}, "you": {},
}

var pathSearchBinaryLikeExt = map[string]struct{}{
	".exe": {}, ".dll": {}, ".dylib": {}, ".so": {}, ".bin": {}, ".pdb": {},
	".obj": {}, ".o": {}, ".class": {}, ".jar": {}, ".wasm": {},
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".ico": {},
	".pdf": {}, ".zip": {}, ".gz": {}, ".7z": {}, ".rar": {},
}

func isPathSearchBinaryLikeBase(name string) bool {
	_, ok := pathSearchBinaryLikeExt[strings.ToLower(filepath.Ext(name))]
	return ok
}

// IsBinaryLikeFileName reports extensions we should not auto-open as editor text (executables, archives, images).
func IsBinaryLikeFileName(name string) bool {
	return isPathSearchBinaryLikeBase(name)
}

// fragmentRefersToBasename reports whether a spoken query is really about this path segment
// (e.g. "find the evade directory" → base "Evade").
func fragmentRefersToBasename(baseName, fragment string) bool {
	baseName = strings.TrimSpace(baseName)
	fragment = strings.TrimSpace(fragment)
	if baseName == "" || fragment == "" {
		return false
	}
	if strings.EqualFold(baseName, fragment) {
		return true
	}
	if NormalizePathTokenForMatch(baseName) == NormalizePathTokenForMatch(fragment) {
		return true
	}
	lowerB := strings.ToLower(baseName)
	for _, w := range strings.Fields(strings.ToLower(fragment)) {
		if len(w) < 2 {
			continue
		}
		if _, stop := fileSearchQueryStopWords[w]; stop {
			continue
		}
		if lowerB == w {
			return true
		}
		if NormalizePathTokenForMatch(baseName) == NormalizePathTokenForMatch(w) {
			return true
		}
	}
	return false
}

// prependWorkspaceRootIfBasenameMatches adds the workspace folder itself when the walk never lists
// path==root but the user is clearly asking for that folder (e.g. repo "Evade" + "Evade Exe.exe" inside).
func prependWorkspaceRootIfBasenameMatches(root, fragment string, matches []PathMatch) []PathMatch {
	root = filepath.Clean(strings.TrimSpace(root))
	fragment = strings.TrimSpace(fragment)
	if root == "" || fragment == "" {
		return matches
	}
	base := filepath.Base(root)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return matches
	}
	if !fragmentRefersToBasename(base, fragment) {
		return matches
	}
	for _, m := range matches {
		if strings.EqualFold(filepath.Clean(m.Path), root) {
			return matches
		}
	}
	return append([]PathMatch{{Path: root, IsDir: true}}, matches...)
}

func pathFragmentMatchScore(m PathMatch, fragment string) int {
	score := 0
	base := filepath.Base(m.Path)
	depth := strings.Count(filepath.ToSlash(filepath.Clean(m.Path)), "/")
	if m.IsDir {
		score += 1000
		if fragmentRefersToBasename(base, fragment) {
			score += 500
		}
	} else {
		score += 100
		if isPathSearchBinaryLikeBase(base) {
			score -= 400
		}
		if fragmentRefersToBasename(base, fragment) {
			score += 200
		}
	}
	score -= depth * 3
	return score
}

func rankPathFragmentMatches(matches []PathMatch, fragment string) {
	sort.SliceStable(matches, func(i, j int) bool {
		si := pathFragmentMatchScore(matches[i], fragment)
		sj := pathFragmentMatchScore(matches[j], fragment)
		if si != sj {
			return si > sj
		}
		return matches[i].Path < matches[j].Path
	})
}

func uniqueSortedPathMatchesCap(items []PathMatch, max int) []PathMatch {
	if len(items) == 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	out := make([]PathMatch, 0, max)
	var last string
	for _, it := range items {
		if it.Path == "" {
			continue
		}
		if len(out) > 0 && it.Path == last {
			continue
		}
		out = append(out, it)
		last = it.Path
		if len(out) >= max {
			break
		}
	}
	return out
}
