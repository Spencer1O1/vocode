package app

import "strings"

type utteranceWindow struct {
	utterances    []string
	maxUtterances int
	maxChars      int
}

func newUtteranceWindow(maxUtterances int, maxChars int) *utteranceWindow {
	if maxUtterances <= 0 {
		maxUtterances = 4
	}
	if maxChars <= 0 {
		maxChars = 1200
	}
	return &utteranceWindow{
		utterances:    make([]string, 0, maxUtterances),
		maxUtterances: maxUtterances,
		maxChars:      maxChars,
	}
}

func (w *utteranceWindow) AddUtterance(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	w.utterances = append(w.utterances, text)
	for len(w.utterances) > w.maxUtterances {
		w.utterances = w.utterances[1:]
	}
	w.trimToMaxChars()
}

func (w *utteranceWindow) PreviousText() string {
	return strings.TrimSpace(strings.Join(w.utterances, " "))
}

func (w *utteranceWindow) trimToMaxChars() {
	for len(w.utterances) > 0 {
		combined := strings.TrimSpace(strings.Join(w.utterances, " "))
		if len(combined) <= w.maxChars {
			return
		}
		w.utterances = w.utterances[1:]
	}
}
