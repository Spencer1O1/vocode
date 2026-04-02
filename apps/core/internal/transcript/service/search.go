package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	"vocoding.net/vocode/v2/apps/core/internal/workspace"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// SearchLikeQueryFromText extracts a literal ripgrep query when the utterance is clearly workspace search.
func SearchLikeQueryFromText(text string) (string, bool) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", false
	}
	lower := strings.ToLower(t)
	prefixes := []string{
		"search for ",
		"find ",
		"search ",
		"where is ",
		"where's ",
		"locate ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			q := strings.TrimSpace(t[len(p):])
			if q == "" {
				return "", false
			}
			return q, true
		}
	}
	return "", false
}

// FileSearchLikeQueryFromText extracts a query when the utterance is clearly a file-path search (not in-file text search).
func FileSearchLikeQueryFromText(text string) (string, bool) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", false
	}
	lower := strings.ToLower(t)
	prefixes := []string{
		"find file ",
		"find files ",
		"file named ",
		"open file ",
		"show file ",
		"locate file ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			q := strings.TrimSpace(t[len(p):])
			if q == "" {
				return "", false
			}
			return q, true
		}
	}
	return "", false
}

type rgHit struct {
	Path    string
	Line0   int
	Char0   int
	Len     int
	Preview string
}

func rgBinary() string {
	if p := strings.TrimSpace(os.Getenv("VOCODE_RG_BIN")); p != "" {
		return p
	}
	return "rg"
}

var rgLineRe = regexp.MustCompile(`^(.*):(\d+):(\d+):(.*)$`)

func rgSearch(root, query string, maxHits int) ([]rgHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if maxHits <= 0 {
		maxHits = 20
	}

	cmd := exec.Command(rgBinary(), "--column", "-n", "--fixed-strings", query, root)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil && stdout.Len() == 0 {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return nil, nil
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}

	out := make([]rgHit, 0, maxHits)
	sc := bufio.NewScanner(bytes.NewReader(stdout.Bytes()))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		m := rgLineRe.FindStringSubmatch(line)
		if len(m) != 5 {
			continue
		}
		path := filepath.Clean(strings.TrimSpace(m[1]))
		ln0 := atoiSafe(m[2])
		col0 := atoiSafe(m[3])
		if ln0 <= 0 || col0 <= 0 {
			continue
		}
		out = append(out, rgHit{
			Path:    path,
			Line0:   ln0 - 1,
			Char0:   col0 - 1,
			Len:     len(query),
			Preview: strings.TrimSpace(m[4]),
		})
		if len(out) >= maxHits {
			break
		}
	}
	return out, nil
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func rgHitsToCompletionHits(hits []rgHit) []struct {
	Path      string `json:"path"`
	Line      int64  `json:"line"`
	Character int64  `json:"character"`
	Preview   string `json:"preview"`
} {
	if len(hits) == 0 {
		return []struct {
			Path      string `json:"path"`
			Line      int64  `json:"line"`
			Character int64  `json:"character"`
			Preview   string `json:"preview"`
		}{}
	}
	out := make([]struct {
		Path      string `json:"path"`
		Line      int64  `json:"line"`
		Character int64  `json:"character"`
		Preview   string `json:"preview"`
	}, 0, len(hits))
	for _, h := range hits {
		out = append(out, struct {
			Path      string `json:"path"`
			Line      int64  `json:"line"`
			Character int64  `json:"character"`
			Preview   string `json:"preview"`
		}{
			Path:      h.Path,
			Line:      int64(h.Line0),
			Character: int64(h.Char0),
			Preview:   h.Preview,
		})
	}
	return out
}

func (s *Service) searchFromText(params protocol.VoiceTranscriptParams, text string, vs *session.VoiceSession) (protocol.VoiceTranscriptCompletion, bool, string) {
	q, ok := SearchLikeQueryFromText(text)
	if !ok {
		return protocol.VoiceTranscriptCompletion{}, false, ""
	}
	return s.searchFromQuery(params, q, vs)
}

func (s *Service) searchFromQuery(params protocol.VoiceTranscriptParams, q string, vs *session.VoiceSession) (protocol.VoiceTranscriptCompletion, bool, string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return protocol.VoiceTranscriptCompletion{}, false, ""
	}
	root := workspace.EffectiveWorkspaceRoot(params.WorkspaceRoot, params.ActiveFile)
	root = strings.TrimSpace(root)
	if root == "" {
		return protocol.VoiceTranscriptCompletion{
			Success:           false,
			Summary:           "",
			TranscriptOutcome: "",
		}, true, "search requires workspaceRoot or activeFile"
	}

	hits, err := rgSearch(root, q, 20)
	if err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "search failed: " + err.Error()
	}

	wireHits := rgHitsToCompletionHits(hits)
	if len(hits) == 0 {
		// Non-nil empty slice clears selection state.
		return protocol.VoiceTranscriptCompletion{
			Success:           true,
			Summary:           fmt.Sprintf("no matches for %q", q),
			TranscriptOutcome: "selection",
			UiDisposition:     "hidden",
			SearchResults:     wireHits,
			ActiveSearchIndex: nil,
		}, true, ""
	}

	var z int64 = 0
	// For compatibility with the extension UX: show hits in the selection panel.
	res := protocol.VoiceTranscriptCompletion{
		Success:           true,
		Summary:           fmt.Sprintf("found %d matches for %q", len(hits), q),
		TranscriptOutcome: "selection",
		UiDisposition:     "hidden",
		SearchResults:     wireHits,
		ActiveSearchIndex: &z,
	}

	// Apply first-hit navigation immediately (parity with daemon UX).
	if s.hostApply == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "daemon has directives but no host apply client is configured"
	}
	first := hits[0]
	dirs := hitNavigateDirectives(first.Path, first.Line0, first.Char0, first.Len)
	batchID := newApplyBatchID()
	if vs != nil {
		vs.PendingDirectiveApply = &session.DirectiveApplyBatch{ID: batchID, NumDirectives: len(dirs)}
	}
	hostRes, err := s.hostApply.ApplyDirectives(protocol.HostApplyParams{
		ApplyBatchId: batchID,
		ActiveFile:   params.ActiveFile,
		Directives:   dirs,
	})
	if err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "host.applyDirectives failed: " + err.Error()
	}
	if vs != nil && vs.PendingDirectiveApply != nil {
		if err := vs.PendingDirectiveApply.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
			vs.PendingDirectiveApply = nil
			return protocol.VoiceTranscriptCompletion{Success: false}, true, "host apply failed: " + err.Error()
		}
		vs.PendingDirectiveApply = nil
	}

	return res, true, ""
}

func pathsFromRgHits(hits []rgHit) []string {
	if len(hits) == 0 {
		return nil
	}
	raw := make([]string, 0, len(hits))
	for _, h := range hits {
		p := strings.TrimSpace(h.Path)
		if p != "" {
			raw = append(raw, p)
		}
	}
	return listUniqueSortedFiles(raw)
}

const fileSearchMaxPaths = 20
const fileSearchMaxRgLines = 120

// fileSearchFromQuery runs a workspace content search and treats unique matching paths as the file-selection hit list.
func (s *Service) fileSearchFromQuery(params protocol.VoiceTranscriptParams, q string, vs *session.VoiceSession) (protocol.VoiceTranscriptCompletion, bool, string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return protocol.VoiceTranscriptCompletion{}, false, ""
	}
	root := workspace.EffectiveWorkspaceRoot(params.WorkspaceRoot, params.ActiveFile)
	root = strings.TrimSpace(root)
	if root == "" {
		return protocol.VoiceTranscriptCompletion{
			Success:           false,
			Summary:           "",
			TranscriptOutcome: "",
		}, true, "search requires workspaceRoot or activeFile"
	}

	hits, err := rgSearch(root, q, fileSearchMaxRgLines)
	if err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "file search failed: " + err.Error()
	}

	paths := pathsFromRgHits(hits)
	if len(paths) > fileSearchMaxPaths {
		paths = paths[:fileSearchMaxPaths]
	}

	if vs != nil {
		vs.SearchResults = nil
		vs.ActiveSearchIndex = 0
		vs.PendingDirectiveApply = nil
		vs.FileSelectionPaths = paths
		if len(paths) > 0 {
			vs.FileSelectionIndex = 0
			vs.FileSelectionFocus = paths[0]
		} else {
			vs.FileSelectionIndex = 0
			vs.FileSelectionFocus = ""
		}
	}

	if len(paths) == 0 {
		return protocol.VoiceTranscriptCompletion{
			Success:           true,
			Summary:           fmt.Sprintf("no file path matches for %q", q),
			TranscriptOutcome: "file_selection",
			UiDisposition:     "hidden",
		}, true, ""
	}

	if s.hostApply == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "daemon has directives but no host apply client is configured"
	}
	first := paths[0]
	dirs := []protocol.VoiceTranscriptDirective{
		{
			Kind: "navigate",
			NavigationDirective: &protocol.NavigationDirective{
				Kind: "success",
				Action: &protocol.NavigationAction{
					Kind: "open_file",
					OpenFile: &struct {
						Path string `json:"path"`
					}{Path: first},
				},
			},
		},
	}
	batchID := newApplyBatchID()
	if vs != nil {
		vs.PendingDirectiveApply = &session.DirectiveApplyBatch{ID: batchID, NumDirectives: len(dirs)}
	}
	hostRes, err := s.hostApply.ApplyDirectives(protocol.HostApplyParams{
		ApplyBatchId: batchID,
		ActiveFile:   params.ActiveFile,
		Directives:   dirs,
	})
	if err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, true, "host.applyDirectives failed: " + err.Error()
	}
	if vs != nil && vs.PendingDirectiveApply != nil {
		if err := vs.PendingDirectiveApply.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
			vs.PendingDirectiveApply = nil
			return protocol.VoiceTranscriptCompletion{Success: false}, true, "host apply failed: " + err.Error()
		}
		vs.PendingDirectiveApply = nil
	}

	return protocol.VoiceTranscriptCompletion{
		Success:                true,
		Summary:                fmt.Sprintf("found %d path(s) for %q", len(paths), q),
		TranscriptOutcome:      "file_selection",
		UiDisposition:          "hidden",
		FileSelectionFocusPath: first,
	}, true, ""
}

// File list helpers
func listUniqueSortedFiles(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	sort.Strings(paths)
	out := make([]string, 0, len(paths))
	var last string
	for i, p := range paths {
		if i == 0 || p != last {
			out = append(out, p)
			last = p
		}
	}
	return out
}
