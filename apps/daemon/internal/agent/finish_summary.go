package agent

import "fmt"

// MaxFinishSummaryRunes is the maximum allowed rune length for a turn-level "done" summary string.
const MaxFinishSummaryRunes = 8192

// ValidateFinishSummary checks the optional transcript summary for a turn-level finish (wire kind "done").
func ValidateFinishSummary(summary string) error {
	if len([]rune(summary)) > MaxFinishSummaryRunes {
		return fmt.Errorf("finish summary exceeds %d characters", MaxFinishSummaryRunes)
	}
	return nil
}
