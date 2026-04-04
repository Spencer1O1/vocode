package workspaceselectflow

import (
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/transcript/hostdirectives"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// resolveEditRange picks an LSP-style range for a scoped edit from params and file text.
func resolveEditRange(params protocol.VoiceTranscriptParams, fileText string) (startLine, startChar, endLine, endChar int, ok bool) {
	lines := strings.Split(fileText, "\n")
	if len(lines) == 0 {
		return 0, 0, 0, 0, false
	}
	last := len(lines) - 1

	if sel := params.ActiveSelection; sel != nil {
		sl := int(sel.StartLine)
		sc := int(sel.StartChar)
		el := int(sel.EndLine)
		ec := int(sel.EndChar)
		if sl == el && sc == ec {
			// caret-only selection: fall through to symbol / file heuristics
		} else {
			if normalizeRange(lines, &sl, &sc, &el, &ec) {
				return sl, sc, el, ec, true
			}
		}
	}

	if cp := params.CursorPosition; cp != nil && len(params.ActiveFileSymbols) > 0 {
		line := int(cp.Line)
		char := int(cp.Character)
		syms := hostdirectives.DocumentSymbolsFromParams(params)
		if sl, sc, el, ec, hit := hostdirectives.SmallestSymbolContainingRange(syms, line, char); hit {
			return sl, sc, el, ec, true
		}
	}

	// Whole file
	endLine = last
	endChar = LineUTF16Len(lines[last])
	return 0, 0, endLine, endChar, true
}

func normalizeRange(lines []string, sl, sc, el, ec *int) bool {
	if *sl < 0 || *el < 0 || *sl >= len(lines) || *el >= len(lines) {
		return false
	}
	if *sl > *el {
		*sl, *el = *el, *sl
		*sc, *ec = *ec, *sc
	}
	if *sc < 0 {
		*sc = 0
	}
	if *ec < 0 {
		*ec = 0
	}
	maxSC := LineUTF16Len(lines[*sl])
	maxEC := LineUTF16Len(lines[*el])
	if *sc > maxSC {
		*sc = maxSC
	}
	if *ec > maxEC {
		*ec = maxEC
	}
	return true
}

// IsWholeFileRange reports whether (sl,sc)-(el,ec) spans the entire file (0,0 through end of last line).
func IsWholeFileRange(fileText string, sl, sc, el, ec int) bool {
	lines := strings.Split(fileText, "\n")
	if len(lines) == 0 {
		return true
	}
	last := len(lines) - 1
	return sl == 0 && sc == 0 && el == last && ec == LineUTF16Len(lines[last])
}

// RangeForSearchHit maps a ripgrep-style hit (0-based line/byte column, byte length) to an LSP UTF-16 range.
// Start/end columns are exclusive on the end, consistent with [extractRangeText].
func RangeForSearchHit(fileText string, hit session.SearchHit) (sl, sc, el, ec int, ok bool) {
	lines := strings.Split(fileText, "\n")
	if hit.Line < 0 || hit.Line >= len(lines) {
		return 0, 0, 0, 0, false
	}
	sl = hit.Line
	byteSC := hit.Character
	if byteSC < 0 || byteSC > len(lines[sl]) {
		return 0, 0, 0, 0, false
	}
	sc = ByteOffsetToUTF16Col(lines[sl], byteSC)
	length := hit.Len
	if length <= 0 {
		length = 1
	}
	line, byteCol := sl, byteSC
	rem := length
	for {
		if line >= len(lines) {
			return 0, 0, 0, 0, false
		}
		cur := lines[line]
		avail := len(cur) - byteCol
		if rem <= avail {
			endByte := byteCol + rem
			return sl, sc, line, ByteOffsetToUTF16Col(cur, endByte), true
		}
		rem -= avail
		line++
		byteCol = 0
		if rem > 0 {
			rem-- // newline
		}
		if rem == 0 {
			return sl, sc, line, 0, true
		}
	}
}

// extractRangeText slices fileText using LSP line/UTF-16 column ranges (matches VS Code / host selection).
func extractRangeText(fileText string, sl, sc, el, ec int) (string, bool) {
	lines := strings.Split(fileText, "\n")
	if sl < 0 || el < sl || el >= len(lines) {
		return "", false
	}
	if sc < 0 || ec < 0 {
		return "", false
	}
	if sc > LineUTF16Len(lines[sl]) || ec > LineUTF16Len(lines[el]) {
		return "", false
	}
	byteSc := UTF16ColToByteOffset(lines[sl], sc)
	byteEc := UTF16ColToByteOffset(lines[el], ec)
	if byteSc > len(lines[sl]) || byteEc > len(lines[el]) {
		return "", false
	}
	if sl == el {
		return lines[sl][byteSc:byteEc], true
	}
	var b strings.Builder
	b.WriteString(lines[sl][byteSc:])
	for i := sl + 1; i < el; i++ {
		b.WriteByte('\n')
		b.WriteString(lines[i])
	}
	b.WriteByte('\n')
	b.WriteString(lines[el][:byteEc])
	return b.String(), true
}
