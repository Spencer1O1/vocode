package stt

import "testing"

func TestExtractStreamingText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "text field", in: `{"text":"hello world"}`, want: "hello world"},
		{name: "transcript field", in: `{"transcript":"hello there"}`, want: "hello there"},
		{name: "empty", in: `{"message_type":"metadata"}`, want: ""},
		{name: "invalid json", in: `{`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStreamingText([]byte(tt.in))
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
