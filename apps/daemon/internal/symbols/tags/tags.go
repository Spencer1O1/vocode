package tags

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const treeSitterLineSep = " | "

// tree-sitter CLI (see crates/cli/src/tags.rs): "def (row, col) - (row, col) `snippet`"
var treeSitterSpanPrefix = regexp.MustCompile(`^(def|ref)\s+\((\d+),\s*(\d+)\)\s+-\s+\((\d+),\s*(\d+)\)\s+`)

// Tag is one symbol tag for a file from `tree-sitter tags`.
// StartLine/EndLine are 0-based row indices (tree-sitter Point).
// StartCharacter/EndCharacter are 0-based UTF-8 byte offsets within each line (tree-sitter Point.column).
// Line is 1-based (start row + 1) for [symbols.BuildSymbolID].
type Tag struct {
	Name           string
	Path           string
	Kind           string
	Line           int
	IsDefinition   bool
	StartLine      int
	StartCharacter int
	EndLine        int
	EndCharacter   int
}

func (t Tag) HasSpan() bool {
	if t.StartLine == 0 && t.StartCharacter == 0 && t.EndLine == 0 && t.EndCharacter == 0 {
		return false
	}
	if t.EndLine > t.StartLine {
		return true
	}
	if t.EndLine == t.StartLine {
		return t.EndCharacter >= t.StartCharacter
	}
	return false
}

// Contains reports whether (line0, byteCol0) lies inside the tag span (tree-sitter coordinates).
func (t Tag) Contains(line0, byteCol0 int) bool {
	if !t.HasSpan() {
		return false
	}
	if line0 < t.StartLine || line0 > t.EndLine {
		return false
	}
	if line0 == t.StartLine && byteCol0 < t.StartCharacter {
		return false
	}
	if line0 == t.EndLine && byteCol0 > t.EndCharacter {
		return false
	}
	return true
}

// LoadTags runs `tree-sitter tags` and parses stdout lines (current CLI format only).
func LoadTags(treeSitterBin, path string) ([]Tag, error) {
	bin := strings.TrimSpace(treeSitterBin)
	if bin == "" {
		return nil, fmt.Errorf("tags: empty tree-sitter binary path")
	}
	p := strings.TrimSpace(path)
	if p == "" {
		return nil, fmt.Errorf("tags: empty path")
	}
	cmd := exec.Command(bin, "tags", p)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tree-sitter tags failed for %q: %w (%s)", p, err, strings.TrimSpace(stderr.String()))
	}
	out := make([]Tag, 0, 32)
	sc := bufio.NewScanner(bytes.NewReader(stdout.Bytes()))
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.Contains(line, treeSitterLineSep) {
			continue
		}
		if tg, ok := parseTreeSitterCLITagLine(line, p); ok {
			out = append(out, tg)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseTreeSitterCLITagLine(line, filePath string) (Tag, bool) {
	idx := strings.Index(line, treeSitterLineSep)
	if idx < 0 {
		return Tag{}, false
	}
	name := strings.TrimSpace(line[:idx])
	if name == "" {
		return Tag{}, false
	}
	right := strings.TrimSpace(line[idx+len(treeSitterLineSep):])
	tab := strings.IndexByte(right, '\t')
	if tab < 0 {
		return Tag{}, false
	}
	syntaxKind := strings.TrimSpace(right[:tab])
	rest := strings.TrimSpace(right[tab+1:])
	m := treeSitterSpanPrefix.FindStringSubmatch(rest)
	if m == nil {
		return Tag{}, false
	}
	sr, _ := strconv.Atoi(m[2])
	scol, _ := strconv.Atoi(m[3])
	er, _ := strconv.Atoi(m[4])
	ecol, _ := strconv.Atoi(m[5])
	after := strings.TrimPrefix(rest, m[0])
	if !strings.HasPrefix(after, "`") {
		return Tag{}, false
	}
	after = after[1:]
	last := strings.LastIndex(after, "`")
	if last < 0 {
		return Tag{}, false
	}
	_ = after[:last]
	role := m[1]
	tg := Tag{
		Name:           name,
		Path:           filepath.Clean(filePath),
		Kind:           normalizeKind(syntaxKind),
		IsDefinition:   role == "def",
		StartLine:      sr,
		StartCharacter: scol,
		EndLine:        er,
		EndCharacter:   ecol,
		Line:           sr + 1,
	}
	if !tg.HasSpan() {
		return Tag{}, false
	}
	return tg, true
}

// NormalizeKind maps tree-sitter syntax type names to a small normalized vocabulary.
func NormalizeKind(kind string) string {
	return normalizeKind(kind)
}

func normalizeKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	k = strings.TrimPrefix(k, "kind:")
	switch k {
	case "f", "func", "function", "function_definition", "function_declaration", "constructor":
		return "function"
	case "m", "method", "member_function", "member":
		return "method"
	case "c", "class", "class_declaration", "struct":
		return "class"
	case "i", "interface", "trait", "protocol":
		return "interface"
	case "e", "enum":
		return "enum"
	case "t", "type", "type_alias", "typedef":
		return "type"
	case "v", "var", "variable", "field", "property", "member_variable":
		return "variable"
	default:
		return k
	}
}
