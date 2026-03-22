package protocol

type Anchor struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type ReplaceAnchoredBlockAction struct {
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Anchor  Anchor `json:"anchor"`
	NewText string `json:"newText"`
}

type EditAction = ReplaceAnchoredBlockAction

type PingParams struct{}

type PingResult struct {
	Message string `json:"message"`
}

type EditApplyParams struct {
	Instruction string `json:"instruction"`
}

type EditApplyResult struct {
	Actions []EditAction `json:"actions"`
}

type JSONRPCRequest[T any] struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

type JSONRPCResponse[T any] struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Result  T      `json:"result"`
}
