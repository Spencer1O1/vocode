package searchapply

import "strings"

// ContentSearchRgVariants returns literal needles to try with ripgrep --fixed-strings when the
// classifier phrase may not appear verbatim in source (e.g. spoken "delta time" vs identifier deltaTime).
func ContentSearchRgVariants(q string) []string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(q)
	ws := strings.Fields(q)
	if len(ws) == 0 {
		return out
	}
	if len(ws) >= 2 {
		var b strings.Builder
		b.WriteString(strings.ToLower(ws[0]))
		for _, w := range ws[1:] {
			w = strings.TrimSpace(w)
			if w == "" {
				continue
			}
			low := strings.ToLower(w)
			r := []rune(low)
			if len(r) == 0 {
				continue
			}
			b.WriteString(strings.ToUpper(string(r[0])))
			if len(r) > 1 {
				b.WriteString(string(r[1:]))
			}
		}
		add(b.String())
	}
	add(strings.Join(ws, ""))
	return out
}
