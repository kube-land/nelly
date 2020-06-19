package nelly

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
)

// WithTrace handler will trace the requests using OpenTelemetry.
func WithTrace() Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

			tr := global.Tracer("pharmatics/nelly")

			attrs, entries, spanCtx := httptrace.Extract(r.Context(), r)

			r = r.WithContext(correlation.ContextWithMap(r.Context(), correlation.NewMap(correlation.MapUpdate{
				MultiKV: entries,
			})))

			ctx, span := tr.Start(
				trace.ContextWithRemoteSpanContext(r.Context(), spanCtx),
				"trace",
				trace.WithAttributes(attrs...),
			)
			defer span.End()

			span.AddEvent(ctx, "trace nelly middleware request")

			r = r.WithContext(ctx)

			h(w, r, p)
		}
	}

	return fn
}
