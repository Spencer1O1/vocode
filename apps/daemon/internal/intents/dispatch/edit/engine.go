package edit

import (
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Engine holds edit action building state (symbol resolution, etc.).
type Engine struct {
	actionBuilder *ActionBuilder
}

func NewEngine() *Engine {
	return &Engine{actionBuilder: NewActionBuilder()}
}

func (e *Engine) BuildActions(ctx EditExecutionContext, editIntent intents.EditIntent) ([]protocol.EditAction, *EditBuildFailure) {
	return e.actionBuilder.BuildActions(ctx, editIntent)
}
