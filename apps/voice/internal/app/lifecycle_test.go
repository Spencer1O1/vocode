package app

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestStopAndTranscribeFinishConcurrent_NoRace(t *testing.T) {
	a := New(strings.NewReader(""), io.Discard)

	for range 200 {
		_, cancel := context.WithCancel(context.Background())

		a.stateMu.Lock()
		a.running = true
		a.cancel = cancel
		a.stateMu.Unlock()

		done := make(chan struct{})
		go func() {
			a.stopIfRunning()
			close(done)
		}()

		a.markTranscribeFinished()
		<-done

		a.stateMu.Lock()
		if a.running {
			a.stateMu.Unlock()
			t.Fatal("running stayed true after concurrent stop/finish")
		}
		if a.cancel != nil {
			a.stateMu.Unlock()
			t.Fatal("cancel func stayed set after concurrent stop/finish")
		}
		a.stateMu.Unlock()
	}
}
