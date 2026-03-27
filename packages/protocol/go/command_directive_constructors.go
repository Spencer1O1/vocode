package protocol

func NewCommandDirective(command string, args []string, timeoutMs int64) CommandDirective {
	return CommandDirective{
		Command:   command,
		Args:      args,
		TimeoutMs: &timeoutMs,
	}
}
