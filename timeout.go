package nelly

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/julienschmidt/httprouter"

	"github.com/pharmatics/rest-util"
)

var errConnKilled = fmt.Errorf("killing connection/stream because serving request timed out and response had been started")

// WithTimeoutForNonLongRunningRequests handler times out non-long-running
// requests after the duration given by requestTimeout.
func WithTimeoutForNonLongRunningRequests(requestTimeout time.Duration) Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {
		timeoutFunc := func(req *http.Request) (*http.Request, <-chan time.Time, func(), *restutil.Status) {
			// TODO unify this with apiserver.MaxInFlightLimit
			ctx := req.Context()

			ctx, cancel := context.WithCancel(ctx)
			req = req.WithContext(ctx)

			postTimeoutFn := func() {
				cancel()
				requestTerminationsTotal.WithLabelValues(req.Method, req.URL.Path, codeToString(http.StatusGatewayTimeout)).Inc()
			}
			return req, time.After(requestTimeout), postTimeoutFn, restutil.NewFailureStatus(fmt.Sprintf("request did not complete within %s", requestTimeout), restutil.StatusReasonTimeout, nil)
		}
		return withTimeout(h, timeoutFunc)
	}
	return fn
}

type timeoutFunc = func(*http.Request) (req *http.Request, timeout <-chan time.Time, postTimeoutFunc func(), err *restutil.Status)

// withTimeout returns an http.Handler that runs h with a timeout
// determined by timeoutFunc. The new http.Handler calls h.ServeHTTP to handle
// each request, but if a call runs for longer than its time limit, the
// handler responds with a 504 Gateway Timeout error and the message
// provided. (If msg is empty, a suitable default message will be sent.) After
// the handler times out, writes by h to its http.ResponseWriter will return
// http.ErrHandlerTimeout. If timeoutFunc returns a nil timeout channel, no
// timeout will be enforced. recordFn is a function that will be invoked whenever
// a timeout happens.
func withTimeout(h httprouter.Handle, timeoutFunc timeoutFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		r, after, postTimeoutFn, err := timeoutFunc(r)
		if after == nil {
			h(w, r, p)
			return
		}

		// resultCh is used as both errCh and stopCh
		resultCh := make(chan interface{})
		tw := newTimeoutWriter(w)
		go func() {
			defer func() {
				err := recover()
				// do not wrap the sentinel ErrAbortHandler panic value
				if err != nil && err != http.ErrAbortHandler {
					// Same as stdlib http server code. Manually allocate stack
					// trace buffer size to prevent excessively large logs
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err = fmt.Sprintf("%v\n%s", err, buf)
				}
				resultCh <- err
			}()
			h(tw, r, p)
		}()
		select {
		case err := <-resultCh:
			// panic if error occurs; stop otherwise
			if err != nil {
				panic(err)
			}
			return
		case <-after:
			defer func() {
				// resultCh needs to have a reader, since the function doing
				// the work needs to send to it. This is defer'd to ensure it runs
				// ever if the post timeout work itself panics.
				go func() {
					res := <-resultCh
					if res != nil {
						switch t := res.(type) {
						case error:
							utilruntime.HandleError(t)
						default:
							utilruntime.HandleError(fmt.Errorf("%v", res))
						}
					}
				}()
			}()

			postTimeoutFn()
			tw.timeout(err)
		}
	}
}

// extend ResponseWriter interface
type timeoutWriter interface {
	http.ResponseWriter
	timeout(*restutil.Status)
}

func newTimeoutWriter(w http.ResponseWriter) timeoutWriter {
	base := &baseTimeoutWriter{w: w}

	_, notifiable := w.(http.CloseNotifier)
	_, hijackable := w.(http.Hijacker)

	switch {
	case notifiable && hijackable:
		return &closeHijackTimeoutWriter{base}
	case notifiable:
		return &closeTimeoutWriter{base}
	case hijackable:
		return &hijackTimeoutWriter{base}
	default:
		return base
	}
}

type baseTimeoutWriter struct {
	w http.ResponseWriter

	mu sync.Mutex
	// if the timeout handler has timeout
	timedOut bool
	// if this timeout writer has wrote header
	wroteHeader bool
	// if this timeout writer has been hijacked
	hijacked bool
}

func (tw *baseTimeoutWriter) Header() http.Header {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return http.Header{}
	}

	return tw.w.Header()
}

func (tw *baseTimeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if tw.hijacked {
		return 0, http.ErrHijacked
	}

	tw.wroteHeader = true
	return tw.w.Write(p)
}

func (tw *baseTimeoutWriter) Flush() {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return
	}

	if flusher, ok := tw.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (tw *baseTimeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut || tw.wroteHeader || tw.hijacked {
		return
	}

	tw.wroteHeader = true
	tw.w.WriteHeader(code)
}

func (tw *baseTimeoutWriter) timeout(err *restutil.Status) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.timedOut = true

	// The timeout writer has not been used by the inner handler.
	// We can safely timeout the HTTP request by sending by a timeout
	// handler
	if !tw.wroteHeader && !tw.hijacked {
		restutil.ResponseJSON(err, tw.w, http.StatusGatewayTimeout)
	} else {
		// The timeout writer has been used by the inner handler. There is
		// no way to timeout the HTTP request at the point. We have to shutdown
		// the connection for HTTP1 or reset stream for HTTP2.
		//
		// Note from: Brad Fitzpatrick
		// if the ServeHTTP goroutine panics, that will do the best possible thing for both
		// HTTP/1 and HTTP/2. In HTTP/1, assuming you're replying with at least HTTP/1.1 and
		// you've already flushed the headers so it's using HTTP chunking, it'll kill the TCP
		// connection immediately without a proper 0-byte EOF chunk, so the peer will recognize
		// the response as bogus. In HTTP/2 the server will just RST_STREAM the stream, leaving
		// the TCP connection open, but resetting the stream to the peer so it'll have an error,
		// like the HTTP/1 case.
		panic(errConnKilled)
	}
}

func (tw *baseTimeoutWriter) closeNotify() <-chan bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		done := make(chan bool)
		close(done)
		return done
	}

	return tw.w.(http.CloseNotifier).CloseNotify()
}

func (tw *baseTimeoutWriter) hijack() (net.Conn, *bufio.ReadWriter, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.timedOut {
		return nil, nil, http.ErrHandlerTimeout
	}
	conn, rw, err := tw.w.(http.Hijacker).Hijack()
	if err == nil {
		tw.hijacked = true
	}
	return conn, rw, err
}

type closeTimeoutWriter struct {
	*baseTimeoutWriter
}

func (tw *closeTimeoutWriter) CloseNotify() <-chan bool {
	return tw.closeNotify()
}

type hijackTimeoutWriter struct {
	*baseTimeoutWriter
}

func (tw *hijackTimeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return tw.hijack()
}

type closeHijackTimeoutWriter struct {
	*baseTimeoutWriter
}

func (tw *closeHijackTimeoutWriter) CloseNotify() <-chan bool {
	return tw.closeNotify()
}

func (tw *closeHijackTimeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return tw.hijack()
}
