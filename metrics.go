package nelly

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/julienschmidt/httprouter"
)

// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/component-base/metrics/prometheus/version/metrics.go

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nelly_request_total",
			Help: "Counter of nelly middleware requests broken out for each verb, resource, client, and HTTP response contentType and code.",
		},
		[]string{"verb", "resource", "client", "contentType", "code"},
	)

	longRunningRequestGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nelly_longrunning_gauge",
			Help: "Gauge of all active long-running nelly middleware requests broken out by verb, resource. Not all requests are tracked this way.",
		},
		[]string{"verb", "resource"},
	)

	requestLatencies = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "nelly_request_duration_seconds",
			Help: "Response latency distribution in seconds for each verb and resource.",
			Buckets: []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
				1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
		},
		[]string{"verb", "resource"},
	)

	responseSizes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "nelly_response_sizes",
			Help: "Response size distribution in bytes for each verb and resource.",
			// Use buckets ranging from 1000 bytes (1KB) to 10^9 bytes (1GB).
			Buckets: prometheus.ExponentialBuckets(1000, 10.0, 7),
		},
		[]string{"verb", "resource"},
	)

	droppedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nelly_dropped_requests_total",
			Help: "Number of requests dropped with 'Try again later' response",
		},
		[]string{"requestKind"},
	)

	currentInflightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nelly_current_inflight_requests",
			Help: "Maximal number of currently used inflight request limit of this nelly middleware per request kind in last second.",
		},
		[]string{"requestKind"},
	)

	requestTerminationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nelly_request_terminations_total",
			Help: "Number of requests which nelly middleware terminated in self-defense.",
		},
		[]string{"verb", "resource", "code"},
	)
)

// RegisterMetrics registers metrics of all Nelly supported middlewares
func RegisterMetrics() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(longRunningRequestGauge)
	prometheus.MustRegister(requestLatencies)
	prometheus.MustRegister(responseSizes)
	prometheus.MustRegister(droppedRequests)
	prometheus.MustRegister(currentInflightRequests)
	prometheus.MustRegister(requestTerminationsTotal)
}

// WithInstrument handler wraps httprouter.Handle to record prometheus metrics
// Recorded metrics:
//
//
func WithInstrument() Handler {

	RegisterMetrics()
	fn := func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {

			now := time.Now()

			delegate := &responseWriterDelegator{ResponseWriter: w}

			_, cn := w.(http.CloseNotifier)
			_, fl := w.(http.Flusher)
			_, hj := w.(http.Hijacker)
			if cn && fl && hj {
				w = &fancyResponseWriterDelegator{delegate}
			} else {
				w = delegate
			}

			h(w, req, p)

			duration := time.Since(now)
			elapsedMicroseconds := float64(duration / time.Microsecond)

			client := req.UserAgent()
			if len(client) == 0 {
				client = "unknown"
			} else if strings.HasPrefix(client, "Mozilla/") {
				client = "Browser"
			}

			requestCounter.WithLabelValues(req.Method, req.URL.Path, client, delegate.Header().Get("Content-type"), codeToString(delegate.Status())).Inc()
			requestLatencies.WithLabelValues(req.Method, req.URL.Path).Observe(elapsedMicroseconds)

			// We are only interested in response sizes of read requests.
			if req.Method == "GET" {
				responseSizes.WithLabelValues(req.Method, req.URL.Path).Observe(float64(delegate.ContentLength()))
			}
		}
	}
	return fn
}

// ResponseWriterDelegator interface wraps http.ResponseWriter to additionally record content-length, status-code, etc.
type responseWriterDelegator struct {
	http.ResponseWriter

	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func (r *responseWriterDelegator) Status() int {
	return r.status
}

func (r *responseWriterDelegator) ContentLength() int {
	return int(r.written)
}

type fancyResponseWriterDelegator struct {
	*responseWriterDelegator
}

func (f *fancyResponseWriterDelegator) CloseNotify() <-chan bool {
	return f.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (f *fancyResponseWriterDelegator) Flush() {
	f.ResponseWriter.(http.Flusher).Flush()
}

func (f *fancyResponseWriterDelegator) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return f.ResponseWriter.(http.Hijacker).Hijack()
}
