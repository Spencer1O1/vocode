package agent

import (
	"fmt"
	"strings"
)

// ScopedEditResult is the only output of the scoped-edit model call.
// The daemon binds this text to a replace_range edit action for the resolved target range.
type ScopedEditResult struct {
	ReplacementText string
}

func (r ScopedEditResult) Validate() error {
	if r.ReplacementText == "" {
		return fmt.Errorf("scoped edit: replacementText must be non-empty")
	}
	if strings.Contains(r.ReplacementText, "\u0000") {
		return fmt.Errorf("scoped edit: replacementText contains NUL")
	}
	return nil
}

