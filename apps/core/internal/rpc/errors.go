package rpc

import protocol "vocoding.net/vocode/v2/packages/protocol/go"

const (
	codeParseError     = -32700
	codeInvalidParams  = -32602
	codeMethodNotFound = -32601
	codeInternalError  = -32000
)

func NewParseError() *protocol.JSONRPCErrorObject {
	return &protocol.JSONRPCErrorObject{
		Code:    codeParseError,
		Message: "Parse error",
	}
}

func NewInvalidParamsError() *protocol.JSONRPCErrorObject {
	return &protocol.JSONRPCErrorObject{
		Code:    codeInvalidParams,
		Message: "Invalid params",
	}
}

func NewMethodNotFoundError() *protocol.JSONRPCErrorObject {
	return &protocol.JSONRPCErrorObject{
		Code:    codeMethodNotFound,
		Message: "Method not found",
	}
}

func NewInternalError(err error) *protocol.JSONRPCErrorObject {
	message := "Internal error"
	if err != nil {
		message = err.Error()
	}

	return &protocol.JSONRPCErrorObject{
		Code:    codeInternalError,
		Message: message,
	}
}

