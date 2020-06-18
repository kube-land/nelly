package nelly

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// WithCacheControl handler sets the Cache-Control header to "no-cache, private" because all servers are supposed to be protected by authn/authz.
// see https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching#defining_optimal_cache-control_policy
func WithCacheControl() Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
			// Set the cache-control header if it is not already set
			if _, ok := w.Header()["Cache-Control"]; !ok {
				w.Header().Set("Cache-Control", "no-cache, private")
			}
			h(w, req, p)
		}
	}

	return fn
}
