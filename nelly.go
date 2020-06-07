package nelly

import (
	"github.com/julienschmidt/httprouter"
)

// Classic returns a new Chain with the default middleware handlers already
// in the chain.
//
// WithTrace - OpenTelemetry tracing handler
//
// WithInstrument - Prometheus Metrics Handler
//
// WithCacheControl - Cache-Control header handler
//
// WithPanicRecovery - Panic Recovery Handler
//
// WithLogging - Logging Handler
func Classic() Chain {
	return NewChain(
		WithTrace(),
		WithInstrument(),
		WithCacheControl(),
		WithPanicRecovery(),
		WithLogging())
}

// A Handler is a generic function that takes httprouter.Handle
// and retrun httprouter.Handle. It is differnet from the common
// signature of middleware handler that use http.Handler
// because it uses julienschmidt/httprouter instead.
type Handler func(httprouter.Handle) httprouter.Handle

// Chain is a list of httprouter.Handle handlers.
// Chain is effectively immutable:
// once created, it will always hold
// the same set of handlers in the same order.
type Chain struct {
	handlers []Handler
}

// NewChain creates a new chain,
// with the given list of handlers.
func NewChain(handlers ...Handler) Chain {
	return Chain{append(([]Handler)(nil), handlers...)}
}

// Then chains the handlers and returns the final httprouter.Handle.
//     NewChain(m1, m2, m3).Then(h)
// is equivalent to:
//     m1(m2(m3(h)))
// When the request comes in, it will be passed to m1, then m2, then m3
// and finally, the given handler
// (assuming every handlers calls the following one).
func (s Chain) Then(h httprouter.Handle) httprouter.Handle {
	for i := range s.handlers {
		h = s.handlers[len(s.handlers)-1-i](h)
	}

	return h
}

// Append extends a chain, adding the specified handler
// as the last ones in the request flow.
//
// Append returns a new chain, leaving the original one untouched.
func (s Chain) Append(handlers ...Handler) Chain {
	newHandlers := make([]Handler, 0, len(s.handlers)+len(handlers))
	newHandlers = append(newHandlers, s.handlers...)
	newHandlers = append(newHandlers, handlers...)

	return Chain{newHandlers}
}

// Extend extends a chain by adding the specified chain
// as the last one in the request flow.
//
// Extend returns a new chain, leaving the original one untouched.
func (s Chain) Extend(chain Chain) Chain {
	return s.Append(chain.handlers...)
}
