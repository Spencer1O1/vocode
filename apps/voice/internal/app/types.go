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
}
