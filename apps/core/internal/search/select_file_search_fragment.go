package search

import (
	"path/filepath"
	"strings"
)

// NormalizeSelectFileSearchQuery returns a single path segment (file or folder basename) for
// [PathFragmentMatches]. The classifier must output only a filename-style token (no slashes);
// this still strips mistakes such as a full absolute path or "Res/game.js" down to "game.js".
func NormalizeSelectFileSearchQuery(root, q string) string {
	_ = root // search scope is the caller's workspace root; basename does not depend on it
	q = strings.TrimSpace(StripFileSearchSpokenFiller(strings.TrimSpace(q)))
	if q == "" {
		return ""
	}
	slash := filepath.ToSlash(q)
	for len(slash) > 1 && strings.HasSuffix(slash, "/") {
		slash = strings.TrimSuffix(slash, "/")
	}
	native := filepath.FromSlash(slash)
	if native == "" {
		return ""
	}
	native = filepath.Clean(native)
	base := filepath.Base(native)
	if base == "." || base == ".." {
		return ""
	}
	base = TrimSttTrailingSentenceDot(base)
	return filepath.ToSlash(base)
}

// stemPathBase returns the filename without its extension for matching spoken stems (e.g. app from
// app.tsx). If there is no extension, returns base unchanged.
func stemPathBase(base string) string {
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		return ""
	}
	ext := filepath.Ext(base)
	if ext == "" || ext == base {
		return base
	}
	return strings.TrimSuffix(base, ext)
}

// ResolveFileSelectSearchFragment picks the path segment used for [PathFragmentMatches]. When the
// flow classifier echoes the active editor basename or stem (e.g. user says "app" while App.tsx is
// open and the model returns App.tsx or app), prefer the fragment derived from the spoken utterance
// so search lists every path match, not a single file-style hit. When the user said "file"
// ([PreferFilesFromSelectQuery]) and the classifier aligns with the open document, always use the
// utterance token so a file literally named app still gets the same broad "app" search as elsewhere.
func ResolveFileSelectSearchFragment(root, classifierQuery, utterance, activeFile string) string {
	classFrag := NormalizeSelectFileSearchQuery(root, classifierQuery)
	if classFrag == "" {
		return ""
	}
	utterFrag := NormalizeSelectFileSearchQuery(root, StripFileSearchSpokenFiller(strings.TrimSpace(utterance)))
	// User said "… file" and named a bare token, but the classifier returned a dotted filename —
	// search by the spoken stem so we list every app* path, not one extension-specific hit.
	if utterFrag != "" &&
		PreferFilesFromSelectQuery(utterance) &&
		fragmentLooksLikeFileOrExtQuery(classFrag) &&
		!fragmentLooksLikeFileOrExtQuery(utterFrag) {
		return utterFrag
	}
	active := strings.TrimSpace(activeFile)
	if active == "" {
		return classFrag
	}
	if utterFrag == "" {
		return classFrag
	}
	actBase := filepath.Base(filepath.Clean(active))
	if actBase == "" || actBase == "." {
		return classFrag
	}
	actStem := stemPathBase(actBase)
	classifierEchoesOpen := strings.EqualFold(classFrag, actBase) ||
		(actStem != "" && strings.EqualFold(classFrag, actStem))
	if !classifierEchoesOpen {
		return classFrag
	}
	if !strings.EqualFold(utterFrag, classFrag) {
		return utterFrag
	}
	if PreferFilesFromSelectQuery(utterance) {
		return utterFrag
	}
	return classFrag
}
