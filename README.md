# nelly
[![Build Status](https://travis-ci.org/pharmatics/nelly.svg?branch=master)](https://travis-ci.org/pharmatics/nelly)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/pharmatics/nelly)
[![Go Report Card](https://goreportcard.com/badge/github.com/pharmatics/nelly)](https://goreportcard.com/report/github.com/pharmatics/nelly)
[![Coverage](http://gocover.io/_badge/github.com/pharmatics/nelly)](http://gocover.io/github.com/pharmatics/nelly)

A high performance and modern HTTP Middleware for Golang (inspired by Kubernetes API server)

![nelly](logo/nelly.png)

## Introduction

Nelly is a minimal chaining middleware, that is designed to work directly with [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) for its high performance and lightweight implementation. Further, it provides some default and useful middleware handlers. Nelly is inspired by [Kubernetes API server filters](https://github.com/kubernetes/apiserver/tree/master/pkg/server/filters) and the chaining middleware [Alice](https://github.com/justinas/alice) which uses the traditional `net/http` handlers.

A list of supported handlers which are recommended to be used in the following order if they are chained:

* [`WithPanicRecovery`](#recovery) - Panic recovery handler
* [`WithLogging`](#logging) - Logging handler for requests and responses
* [`WithInstrument`](#metrics) -  Prometheus metrics handler for requests and responses
* [`WithCacheControl`](#cache-control) - Cache-Control header handler to set `"no-cache, private"`
* [`WithTimeoutForNonLongRunningRequests`](#timeout) - Timeout handler for non-long running requests
* [`WithCORS`](#cors) -  CORS (Cross-Origin Resource Sharing) headers handler
* [`WithRequiredHeaders`](#headers) - Headers handler to check missing headers
* [`WithRequiredHeaderValues`](#headers) - Headers handler to check invalid headers values
* [`WithAuthSigningMethodHS256`](#authentication) - Authentication handler to validate JWT token using HS256 algorithm
* [`WithAuthSigningMethodRS256`](#authentication) - Authentication handler to validate JWT token using RS256 algorithm

## Getting Started

A chian is an immutable list of a middleware handlers. Its handlers have an order where the request pass through the chain. The middleware handlers have the form 

```go
// nelly package
type Handler func(httprouter.Handle) httprouter.Handle
```

where the `httprouter` handle has the form

```go
// httprouter package
type Handle func(http.ResponseWriter, *http.Request, Params)
```

To write a new middleware handler that could be chained to other middleware handlers as in the following example:

```go
func someHandler(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// middleware handler implementation goes here
		h(w, r, p)
	}
}
```

Further, any of the default middleware handlers that is provided by `nelly` could be used to create a new chain. To create a new chain form a set of handlers:

```go
chain := nelly.NewChain(someHandler, otherHanlder, ...)
```

To wrap your handler `appHandler` (`httprouter.Handle`) with the created chain:

```go
func appHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// httprouter.Handle implementation goes here
	return
}

chainedHandler := chain.Then(appHandler)
```

Then you can use the created `chainedHandler` with `httprouter`

```go
// create new router
router := httprouter.New()

// use chainedHandler (httprouter.Handle)
router.GET("/", chainedHandler)
log.Fatal(http.ListenAndServe(":8080", router))
```

The requests will pass `someHandler` first, then `otherHanlder` till the end of the set of the passed handlers to `NewChain` in the same order, and finally to `appHandler` which is equivalent to:

```go
someHandler(otherHanlder(...(appHandler))
```

For more flexible usage of the created `chain`, you can use `Append(handler)` or `Extend(chain)`:

* `Append` - will return a new chain, leaving the original one untouched, and adds the specified middleware handlers as the last ones in the request flow of the original chain.

* `Extend` - will return a new chain, leaving the original one untouched, and adds the specified chain as the last one in the request flow of the original chain.

```go
// Create new chian
chain := nelly.NewChain(handler_A, handler_B)

// Append to the chain some handlers
chain_1 := chain.Append(handler_C, handler_D)

// Create another chain
chain_2 := nelly.NewChain(handler_E, handler_F)

// Create new chian by chaining chain1 with chain2
newChain := chain1.Extend(chain2)

// wrap the appHandler with the created chain
chainedHandler := chain2.Then(appHandler)
```

In previous example using `chainedHandler` in `httprouter` will pass the requests as follow

`handler_A -> handler_B -> handler_C -> handler_D -> handler_E -> handler_F -> appHandler`

## Default Handlers

The `Classic()` version of `nelly` returns a new Chain with some default middleware handlers already in the chain with the following order:

`WithPanicRecovery() -> WithLogging() -> WithInstrument() -> WithCacheControl()`

To use classic version:

```go
// create classic chain
classicChain := nelly.Classic()

// extend the classic chain with some default chain (WithCORS)
chian := classicChain.Append(WithCORS(opts), ...)
```

Using any of the default handlers which implements middleware handler is recommended in the following order:

### Recovery

## Tracing (OpenTelemetry)

Nelly had a support for tracing using [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) which has been deprecated in the favour of `othttp` (OpenTelemetry HTTP Handler) which only support the traditional `net/http` handler. Unfortunately, it is not possible to use `othttp` directly with nelly middleware, but it could be used with `julienschmidt/httprouter` as follow:

```go
// import "go.opentelemetry.io/otel/instrumentation/othttp"

// create new router
router := httprouter.New()

// use chainedHandler (httprouter.Handle)
router.GET("/", chainedHandler)

// wrap httprouter with OpenTelemetry http
otWrap := othttp.NewHandler(router, "server")

log.Fatal(http.ListenAndServe(":8080", otWrap))
```

The router will be wrapped with `othttp` handler which support the traditional `net/http` Handler.

# Todo:

- [ ] Improve metrics handler
- [ ] Improve timeout for long running requests
