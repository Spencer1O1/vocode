package agent

import (
	"testing"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

func TestClassifierResultValidate(t *testing.T) {
	t.Parallel()
	if err := (ClassifierResult{Flow: flows.Root, Route: "select"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ClassifierResult{Flow: flows.Root, Route: "bogus"}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (ClassifierResult{Flow: flows.Select, Route: "select_control"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ClassifierResult{Flow: flows.SelectFile, Route: "open"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
