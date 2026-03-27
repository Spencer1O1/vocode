package protocol

func NewEditDirectiveSuccess(actions []EditAction) EditDirective {
	return EditDirective{
		Kind:    "success",
		Actions: actions,
	}
}

func NewEditDirectiveNoop(reason string) EditDirective {
	return EditDirective{
		Kind:   "noop",
		Reason: reason,
	}
}
