package nelly

import (
	"github.com/julienschmidt/httprouter"
	"github.com/pharmatics/rest-util"
	"net/http"
)

// WithRequiredHeaders handler checks if a list of headers are set on requests. If
// any header from the list doesn't exist, the handler will return StatusBadRequest
func WithRequiredHeaders(requiredHeaders []string) Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {

		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

			var missing []string

			for _, h := range requiredHeaders {
				if r.Header.Get(h) == "" {
					missing = append(missing, h)
				}
			}

			if len(missing) != 0 {
				statusErr := restutil.ErrorWithDetails("Missing headers", restutil.StatusReasonBadRequest, missing)
				restutil.ResponseJSON(statusErr, w, statusErr.Code)
				return
			}

			h(w, r, p)
		}
	}

	return fn

}

// WithRequiredHeaderValues handler checks if a map of headers are set on requests.
// If any header from the map doesn't exist or don't equal the values in
// requiredHeaderValues, the handler will return StatusBadRequest
func WithRequiredHeaderValues(requiredHeaderValues map[string]string) Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

			var invalid []string

			for h, v := range requiredHeaderValues {
				if r.Header.Get(h) != v {
					invalid = append(invalid, h)
				}
			}

			if len(invalid) != 0 {
				statusErr := restutil.ErrorWithDetails("Invalid headers", restutil.StatusReasonBadRequest, invalid)
				restutil.ResponseJSON(statusErr, w, statusErr.Code)
				return
			}

			h(w, r, p)
		}
	}

	return fn
}
