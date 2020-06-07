package nelly

import (
	"net/http"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"

	"github.com/julienschmidt/httprouter"
)

// WithPanicRecovery handler wraps an httprouter.Handle to recover and log panics
func WithPanicRecovery() Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {

		return withPanicRecovery(h, func(w http.ResponseWriter, req *http.Request, err interface{}) {
			if err == http.ErrAbortHandler {
				// honor the http.ErrAbortHandler sentinel panic value:
				//   ErrAbortHandler is a sentinel panic value to abort a handler.
				//   While any panic from ServeHTTP aborts the response to the client,
				//   panicking with ErrAbortHandler also suppresses logging of a stack trace to the server's error log.
				return
			}
			http.Error(w, "This request caused nelly middleware to panic. Look in the logs for details.", http.StatusInternalServerError)
			klog.Errorf("nelly middleware panic'd on %v %v", req.Method, req.RequestURI)
		})
	}

	return fn
}

func withPanicRecovery(handler httprouter.Handle, crashHandler func(http.ResponseWriter, *http.Request, interface{})) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		defer runtime.HandleCrash(func(err interface{}) {
			crashHandler(w, req, err)
		})

		// Dispatch to the internal handler
		handler(w, req, p)
	}
}
