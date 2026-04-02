package workspace

import (
	"path/filepath"
	"strings"
)

// EffectiveWorkspaceRoot mirrors the daemon’s behavior at a basic level:
// - if a workspaceRoot is provided, use it
// - otherwise fall back to the parent directory of the active file
func EffectiveWorkspaceRoot(workspaceRoot, activeFile string) string {
	root := strings.TrimSpace(workspaceRoot)
	if root != "" {
		return filepath.Clean(root)
	}
	active := strings.TrimSpace(activeFile)
	if active == "" {
		return ""
	}
	return filepath.Dir(filepath.Clean(active))
}

