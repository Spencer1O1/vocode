package tags

import (
	"bufio"
	"fmt"
	"os"
	"unicode/utf8"
)

// ByteOffsetForLineAndUTF16Column returns the tree-sitter-style byte column (0-based) within line line0
// that corresponds to utf16CodeUnitIndex UTF-16 code units from the start of that line.
func ByteOffsetForLineAndUTF16Column(absPath string, line0, utf16CodeUnitIndex int) (int, error) {
	if line0 < 0 || utf16CodeUnitIndex < 0 {
		return 0, fmt.Errorf("invalid position")
	}
	f, err := os.Open(absPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for l := 0; sc.Scan(); l++ {
		if l == line0 {
			return utf16PrefixByteLength(sc.Text(), utf16CodeUnitIndex), nil
		}
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("line %d out of range", line0)
}

func utf16PrefixByteLength(line string, wantUTF16 int) int {
	if wantUTF16 <= 0 {
		return 0
	}
	b := 0
	u := 0
	for b < len(line) {
		if u >= wantUTF16 {
			break
		}
		r, sz := utf8.DecodeRuneInString(line[b:])
		if r == utf8.RuneError && sz == 1 {
			b++
			continue
		}
		var add int
		if r > 0xFFFF {
			add = 2
		} else {
			add = 1
		}
		if u+add > wantUTF16 {
			break
		}
		u += add
		b += sz
	}
	return b
}
