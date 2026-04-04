package workspaceselectflow

import "unicode/utf8"

// LineUTF16Len returns the LSP / VS Code "character" extent of a single line (UTF-16 code units).
func LineUTF16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// UTF16ColToByteOffset maps an LSP 0-based UTF-16 column within s to a byte offset for Go string slicing.
func UTF16ColToByteOffset(s string, col int) int {
	if col <= 0 {
		return 0
	}
	utf16 := 0
	i := 0
	for i < len(s) {
		r, w := utf8.DecodeRuneInString(s[i:])
		add := 1
		if r >= 0x10000 {
			add = 2
		}
		if utf16+add > col {
			return i
		}
		utf16 += add
		i += w
		if utf16 == col {
			return i
		}
	}
	return len(s)
}

// ByteOffsetToUTF16Col maps a byte offset within line s to an LSP UTF-16 column (0-based).
func ByteOffsetToUTF16Col(s string, byteOff int) int {
	if byteOff <= 0 {
		return 0
	}
	if byteOff > len(s) {
		byteOff = len(s)
	}
	return LineUTF16Len(s[:byteOff])
}
