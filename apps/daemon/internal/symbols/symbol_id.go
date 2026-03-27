package symbols

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const symbolIDVersion = "v1"

// BuildSymbolID encodes a stable, transport-safe identifier for a symbol ref.
func BuildSymbolID(ref SymbolRef) string {
	path := base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(ref.Path)))
	kind := base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(ref.Kind)))
	name := base64.RawURLEncoding.EncodeToString([]byte(strings.TrimSpace(ref.Name)))
	line := strconv.Itoa(ref.Line)
	return strings.Join([]string{symbolIDVersion, path, line, kind, name}, "|")
}

// ParseSymbolID decodes an ID produced by BuildSymbolID.
func ParseSymbolID(id string) (SymbolRef, error) {
	parts := strings.Split(strings.TrimSpace(id), "|")
	if len(parts) != 5 || parts[0] != symbolIDVersion {
		return SymbolRef{}, fmt.Errorf("invalid symbol id format")
	}
	pathBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return SymbolRef{}, fmt.Errorf("invalid symbol id path: %w", err)
	}
	line, err := strconv.Atoi(parts[2])
	if err != nil || line < 0 {
		return SymbolRef{}, fmt.Errorf("invalid symbol id line")
	}
	kindBytes, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return SymbolRef{}, fmt.Errorf("invalid symbol id kind: %w", err)
	}
	nameBytes, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return SymbolRef{}, fmt.Errorf("invalid symbol id name: %w", err)
	}
	return SymbolRef{
		ID:   strings.TrimSpace(id),
		Path: string(pathBytes),
		Line: line,
		Kind: string(kindBytes),
		Name: string(nameBytes),
	}, nil
}
