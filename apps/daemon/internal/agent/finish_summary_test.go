package agent

import (
	"strings"
	"testing"
)

func TestValidateFinishSummary(t *testing.T) {
	t.Parallel()
	if err := ValidateFinishSummary("Applied edits and ran tests."); err != nil {
		t.Fatalf("expected valid summary: %v", err)
	}
}

func TestValidateFinishSummaryTooLong(t *testing.T) {
	t.Parallel()
	s := strings.Repeat("x", MaxFinishSummaryRunes+1)
	if err := ValidateFinishSummary(s); err == nil {
		t.Fatal("expected error for oversized finish summary")
	}
}
