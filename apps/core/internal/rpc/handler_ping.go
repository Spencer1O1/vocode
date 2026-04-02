package rpc

import (
	"encoding/json"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func NewPingHandler() Handler {
	return func(req protocol.JSONRPCRequest[json.RawMessage]) (any, *protocol.JSONRPCErrorObject) {
		_, rpcErr := DecodeParams[protocol.PingParams](req.Params)
		if rpcErr != nil {
			return nil, rpcErr
		}

		return protocol.PingResult{
			Message: "pong",
		}, nil
	}
}

