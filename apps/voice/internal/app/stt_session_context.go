package app

import "strings"

// elevenLabsSTTSessionContextPrefix is sent as ElevenLabs previous_text on the first audio chunk of
// each WebSocket session (combined with rolling committed utterances). It biases the model toward
// programming dictation; not user-configurable.
const elevenLabsSTTSessionContextPrefix = "I am giving you spoken instructions for programming and editing code."

// elevenLabsPreviousText returns text for input_audio_chunk.previous_text: fixed session context
// plus any rolling committed transcript from this session (see utteranceWindow).
func elevenLabsPreviousText(w *utteranceWindow) string {
	p := strings.TrimSpace(elevenLabsSTTSessionContextPrefix)
	u := strings.TrimSpace(w.PreviousText())
	if u == "" {
		return p
	}
	return p + " " + u
}
