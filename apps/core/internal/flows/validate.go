package flows

import (
	"fmt"
	"strings"
)

// ValidateRoute returns an error if route is empty or not defined for the flow.
func ValidateRoute(flow ID, route string) error {
	r := strings.TrimSpace(route)
	if r == "" {
		return fmt.Errorf("flow classifier: empty route")
	}
	for _, id := range SpecFor(flow).RouteIDs() {
		if id == r {
			return nil
		}
	}
	return fmt.Errorf("flow classifier: unknown route %q for flow %q", r, flow)
}
