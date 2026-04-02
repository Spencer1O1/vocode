package workspace

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// ListWorkspaceFiles returns a stable, workspace-jail-safe list of file paths under root.
// It includes regular files only (no directories).
func ListWorkspaceFiles(root string, maxFiles int) ([]string, error) {
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "" {
		return nil, nil
	}
	if maxFiles <= 0 {
		maxFiles = 2000
	}

	out := make([]string, 0, 256)
	// WalkDir keeps allocations low and is easy to stop via an error.
	stopErr := fs.ErrClosed
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		out = append(out, filepath.Clean(path))
		if len(out) >= maxFiles {
			return stopErr
		}
		return nil
	})

	// Ignore stop sentinel.
	if err != nil && err != stopErr {
		return nil, err
	}

	sort.Strings(out)
	return out, nil
}

