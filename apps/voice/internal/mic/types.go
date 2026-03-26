package mic

// StartParams configures native microphone capture.
// The current STT pipeline assumes mono, so Channels should be 1.
type StartParams struct {
	SampleRateHz int
	Channels     int
}
