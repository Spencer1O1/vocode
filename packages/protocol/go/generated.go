// AUTO-GENERATED. DO NOT EDIT.

package protocol

type Anchor struct {
	Before string `json:"before"`
	After string `json:"after"`
}

type ReplaceBetweenAnchorsAction struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
	Anchor Anchor `json:"anchor"`
	NewText string `json:"newText"`
}

type EditAction = ReplaceBetweenAnchorsAction

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

type JSONRPCErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JSONRPCErrorResponse struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      *int64            `json:"id"`
	Error   JSONRPCErrorObject `json:"error"`
}

