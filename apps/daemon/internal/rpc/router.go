package rpc

import (
	"encoding/json"
	"log"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

type Handler func(
	req protocol.JSONRPCRequest[json.RawMessage],
) (any, *protocol.JSONRPCErrorObject)

type Router struct {
	logger   *log.Logger
	handlers map[string]Handler
}

func NewRouter(logger *log.Logger) *Router {
	return &Router{
		logger:   logger,
		handlers: make(map[string]Handler),
	}
}

func (r *Router) Register(method string, handler Handler) {
	r.handlers[method] = handler
}

func (r *Router) Handle(
	req protocol.JSONRPCRequest[json.RawMessage],
) (any, *protocol.JSONRPCErrorObject) {
	handler, ok := r.handlers[req.Method]
	if !ok {
		return nil, NewMethodNotFoundError()
	}

	return handler(req)
}
