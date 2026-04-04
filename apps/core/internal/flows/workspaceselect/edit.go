package workspaceselectflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// HandleEdit runs a scoped edit: resolve range from selection/cursor/symbols, ask the model for replacement text, apply replace_range.
func HandleEdit(deps *SelectionDeps, params protocol.VoiceTranscriptParams, vs *session.VoiceSession, text string) (protocol.VoiceTranscriptCompletion, string) {
	instr := strings.TrimSpace(text)
	if instr == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "edit: empty instruction"
	}
	active := strings.TrimSpace(params.ActiveFile)
	if active == "" {
		return protocol.VoiceTranscriptCompletion{Success: false},
			"Open a file in the editor first. Edit changes code in the active editor."
	}
	if deps.ExtensionHost == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "extension host not configured"
	}
	if deps.EditModel == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "edit: no model configured (set VOCODE_AGENT_PROVIDER=openai and API keys)"
	}
	body, err := deps.ExtensionHost.ReadHostFile(active)
	if err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "read file: " + err.Error()
	}
	sl, sc, el, ec, ok := resolveEditRange(params, body)
	if !ok {
		return protocol.VoiceTranscriptCompletion{Success: false}, "edit: could not resolve target range"
	}
	targetText, ok := extractRangeText(body, sl, sc, el, ec)
	if !ok {
		return protocol.VoiceTranscriptCompletion{Success: false}, "edit: invalid target range"
	}

	sum := sha256.Sum256([]byte(targetText))
	fp := hex.EncodeToString(sum[:])

	modelOut, err := callScopedEditModel(context.Background(), deps.EditModel, instr, active, sl, sc, el, ec, targetText, body)
	if err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "edit model: " + err.Error()
	}

	if deps.HostApply == nil || deps.NewBatchID == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply client not configured"
	}
	batchID := deps.NewBatchID()

	filteredImports := filterNewImportLines(body, modelOut.ImportLines)
	importBlock := importBlockForInsert(filteredImports)
	insertLine := importInsertLine(strings.Split(body, "\n"), active)
	lineOff := linesAddedByImportBlock(importBlock)

	sl2, sc2, el2, ec2 := sl, sc, el, ec
	if importBlock != "" {
		sl2, sc2, el2, ec2 = shiftRangeAfterImportInsert(sl, sc, el, ec, insertLine, lineOff)
	}

	actions := make([]protocol.EditAction, 0, 2)
	if importBlock != "" {
		actions = append(actions, protocol.EditAction{
			Kind:           "replace_range",
			Path:           active,
			NewText:        importBlock,
			ExpectedSha256: emptySHA256,
			Range: &struct {
				StartLine int64 `json:"startLine"`
				StartChar int64 `json:"startChar"`
				EndLine   int64 `json:"endLine"`
				EndChar   int64 `json:"endChar"`
			}{
				StartLine: int64(insertLine),
				StartChar: 0,
				EndLine:   int64(insertLine),
				EndChar:   0,
			},
			EditId: "vocode-imports-" + batchID,
		})
	}
	actions = append(actions, protocol.EditAction{
		Kind:           "replace_range",
		Path:           active,
		NewText:        modelOut.ReplacementText,
		ExpectedSha256: fp,
		Range: &struct {
			StartLine int64 `json:"startLine"`
			StartChar int64 `json:"startChar"`
			EndLine   int64 `json:"endLine"`
			EndChar   int64 `json:"endChar"`
		}{
			StartLine: int64(sl2),
			StartChar: int64(sc2),
			EndLine:   int64(el2),
			EndChar:   int64(ec2),
		},
		EditId: "vocode-edit-" + batchID,
	})

	dir := []protocol.VoiceTranscriptDirective{
		{
			Kind: "edit",
			EditDirective: &protocol.EditDirective{
				Kind:    "success",
				Actions: actions,
			},
		},
	}

	shouldOrganize := len(filteredImports) > 0 && modelOut.OrganizeImports
	if shouldOrganize {
		dir = append(dir, protocol.VoiceTranscriptDirective{
			Kind: "code_action",
			CodeActionDirective: &struct {
				Path   string `json:"path"`
				Range  *struct {
					StartLine int64 `json:"startLine"`
					StartChar int64 `json:"startChar"`
					EndLine   int64 `json:"endLine"`
					EndChar   int64 `json:"endChar"`
				} `json:"range,omitempty"`
				ActionKind             string `json:"actionKind"`
				PreferredTitleIncludes string `json:"preferredTitleIncludes,omitempty"`
			}{
				Path:       active,
				ActionKind: "source.organizeImports",
			},
		})
	}

	pending := &session.DirectiveApplyBatch{ID: batchID, NumDirectives: len(dir)}
	if vs != nil {
		vs.PendingDirectiveApply = pending
	}
	hostRes, err := deps.HostApply.ApplyDirectives(protocol.HostApplyParams{
		ApplyBatchId: batchID,
		ActiveFile:   params.ActiveFile,
		Directives:   dir,
	})
	if err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, "host.applyDirectives failed: " + err.Error()
	}
	if err := pending.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply failed: " + err.Error()
	}
	if vs != nil {
		vs.PendingDirectiveApply = nil
	}

	return protocol.VoiceTranscriptCompletion{
		Success:       true,
		Summary:       "applied edit",
		UiDisposition: "hidden",
	}, ""
}

const scopedEditFullFileContextMaxBytes = 120_000

func truncateScopedEditFileContext(s string) string {
	const max = scopedEditFullFileContextMaxBytes
	if len(s) <= max {
		return s
	}
	s = s[:max]
	for len(s) > 0 && s[len(s)-1]&0xC0 == 0x80 {
		s = s[:len(s)-1]
	}
	return s + "\n…(truncated for model context)…"
}

type scopedEditModelOut struct {
	ReplacementText string
	ImportLines     []string
	OrganizeImports bool
}

func callScopedEditModel(ctx context.Context, m agent.ModelClient, instruction, activeFile string, sl, sc, el, ec int, targetText, fullFile string) (scopedEditModelOut, error) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"replacementText"},
		"properties": map[string]any{
			"replacementText": map[string]any{
				"type":        "string",
				"description": "Replacement for targetText only; must match language, libraries, and idioms evident in targetText and the file path. In React Native/Expo files use RN components and onPress, never HTML tags or onClick.",
			},
			"importLines": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
				"description": `Optional import lines for the host to insert. JS/TS: full lines (e.g. import { foo } from "bar"). ` +
					`Go single import: import "encoding/json". Go grouped import ( ... ): use spec lines only inside the parens (e.g. "errors" or m "example.com/lib"). ` +
					`Omit or [] if none. Do not duplicate lines already in fullFile.`,
			},
			"organizeImports": map[string]any{
				"type":        "boolean",
				"description": `When importLines is non-empty: if true (default), the host runs the editor "organize imports" action after applying (works for TypeScript/JavaScript via tsserver and Go via gopls when installed). Set false to skip.`,
			},
		},
	}
	type targetPayload struct {
		Path      string `json:"path"`
		StartLine int    `json:"startLine"`
		StartChar int    `json:"startChar"`
		EndLine   int    `json:"endLine"`
		EndChar   int    `json:"endChar"`
	}
	userObj := map[string]any{
		"instruction": instruction,
		"activeFile":  activeFile,
		"target": targetPayload{
			Path:      activeFile,
			StartLine: sl,
			StartChar: sc,
			EndLine:   el,
			EndChar:   ec,
		},
		"targetText": targetText,
		"fullFile":   truncateScopedEditFileContext(fullFile),
	}
	userBytes, err := json.MarshalIndent(userObj, "", "  ")
	if err != nil {
		return scopedEditModelOut{}, err
	}
	sys := strings.TrimSpace(`
You are Vocode's scoped edit model. You receive the user's instruction, active file path, a target range, targetText (exact source in that range), and fullFile (entire buffer, possibly truncated at the end).

Output JSON with:
- replacementText: the literal new source for the target range only. Never paste the whole file. Do not put import statements here unless the selection itself is an import block.
- importLines (optional): array of complete new import lines the host will insert into the file's import section (e.g. import { Pressable } from "react-native"). Omit or use [] if nothing new is needed. Never repeat a line that already appears in fullFile.
- organizeImports (optional): when importLines is non-empty, defaults to true — host runs TypeScript/JavaScript "organize imports" after applying; set false to leave order as-is.

Use fullFile for awareness: existing imports, components, hooks, and patterns. Infer language and stack from targetText, fullFile, and path.

replacementText should read as if the same author wrote it: consistent naming and structure with the rest of the file.

No markdown fences or extra keys beyond replacementText, importLines, and organizeImports.
`) + reactNativeExpoRules
	out, err := m.Call(ctx, agent.CompletionRequest{
		System:     sys,
		User:       string(userBytes),
		JSONSchema: schema,
	})
	if err != nil {
		return scopedEditModelOut{}, err
	}
	var parsed struct {
		ReplacementText string   `json:"replacementText"`
		ImportLines     []string `json:"importLines"`
		OrganizeImports *bool    `json:"organizeImports"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		return scopedEditModelOut{}, fmt.Errorf("decode model json: %w", err)
	}
	result := scopedEditModelOut{ReplacementText: parsed.ReplacementText, ImportLines: parsed.ImportLines}
	if len(parsed.ImportLines) > 0 {
		if parsed.OrganizeImports != nil {
			result.OrganizeImports = *parsed.OrganizeImports
		} else {
			result.OrganizeImports = true
		}
	}
	return result, nil
}
