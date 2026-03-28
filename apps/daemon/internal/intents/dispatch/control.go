package dispatch

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/done"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/requestcontext"
)

// controlDispatch turns one validated [intents.ControlIntent] into a [ControlResult]
// (no protocol directives). [doneControl] ignores h and in; [requestContextControl] uses
// [Handler.request], transcript params, and planning context from [HandleInput].
type controlDispatch interface {
	dispatch(h *Handler, in HandleInput) (*ControlResult, error)
}

func controlFor(c *intents.ControlIntent) (controlDispatch, error) {
	switch c.Kind {
	case intents.ControlIntentKindDone:
		return doneControl{}, nil
	case intents.ControlIntentKindRequestContext:
		return requestContextControl{req: c.RequestContext}, nil
	default:
		return nil, fmt.Errorf("unknown control intent kind %q", c.Kind)
	}
}

type doneControl struct{}

func (doneControl) dispatch(_ *Handler, _ HandleInput) (*ControlResult, error) {
	_, err := done.Dispatch()
	if err != nil {
		return nil, err
	}
	return &ControlResult{Done: &DoneResult{}}, nil
}

type requestContextControl struct {
	req *intents.RequestContextIntent
}

func (r requestContextControl) dispatch(h *Handler, in HandleInput) (*ControlResult, error) {
	updated, err := requestcontext.Dispatch(h.request, in.Params, in.TurnCtx, r.req)
	if err != nil {
		return nil, err
	}
	return &ControlResult{
		Fulfilled: &RequestContextFulfilled{PlanningContext: updated},
	}, nil
}
