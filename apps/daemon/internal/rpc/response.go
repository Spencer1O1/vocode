package rpc

import protocol "vocoding.net/vocode/v2/packages/protocol/go"

const jsonRPCVersion = "2.0"

func NewSuccessResponse(id int64, result any) protocol.JSONRPCResponse[any] {
	return protocol.JSONRPCResponse[any]{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id *int64, err *protocol.JSONRPCErrorObject) protocol.JSONRPCErrorResponse {
	return protocol.JSONRPCErrorResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error:   *err,
	}
}
