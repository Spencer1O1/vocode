package rpc

import (
	"bytes"
	"encoding/json"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func DecodeParams[T any](raw json.RawMessage) (T, *protocol.JSONRPCErrorObject) {
	var params T

	if len(bytes.TrimSpace(raw)) == 0 {
		return params, nil
	}

	if err := json.Unmarshal(raw, &params); err != nil {
		rpcErr := NewInvalidParamsError()
		return params, rpcErr
	}

	return params, nil
}
