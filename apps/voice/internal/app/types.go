package app

type Request struct {
	Type string `json:"type"`
}

type Event struct {
	Type    string `json:"type"`
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Text    string `json:"text,omitempty"`
	// Committed indicates whether this transcript is a final/committed hypothesis.
	// When omitted, the event is considered backwards-compatible.
	Committed *bool `json:"committed,omitempty"`
}
