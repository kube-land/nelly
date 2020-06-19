package nelly

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog"

	"github.com/julienschmidt/httprouter"
)

// CORSOpts is the configuration that will be used by WithCORS
type CORSOpts struct {
	AllowedOriginPatterns []string
	AllowedMethods        []string
	AllowedHeaders        []string
	ExposedHeaders        []string
	AllowCredentials      bool
}

// WithCORS handler is a simple CORS implementation that wraps an httprouter.Handle.
// Pass nil for allowedMethods and allowedHeaders to use the defaults. If allowedOriginPatterns
// is empty, no CORS support is installed.
func WithCORS(opts CORSOpts) Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {
		return withCORS(h,
			opts.AllowedOriginPatterns,
			opts.AllowedMethods,
			opts.AllowedHeaders,
			opts.ExposedHeaders,
			strconv.FormatBool(opts.AllowCredentials))
	}

	return fn
}

func withCORS(handler httprouter.Handle, allowedOriginPatterns []string, allowedMethods []string, allowedHeaders []string, exposedHeaders []string, allowCredentials string) httprouter.Handle {

	if len(allowedOriginPatterns) == 0 {
		return handler
	}

	allowedOriginPatternsREs := allowedOriginRegexps(allowedOriginPatterns)

	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, re := range allowedOriginPatternsREs {
				if allowed = re.MatchString(origin); allowed {
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// Set defaults for methods and headers if nothing was passed
				if allowedMethods == nil {
					allowedMethods = []string{"POST", "GET", "OPTIONS", "PUT", "DELETE", "PATCH"}
				}
				if allowedHeaders == nil {
					allowedHeaders = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With", "If-Modified-Since"}
				}
				if exposedHeaders == nil {
					exposedHeaders = []string{"Date"}
				}

				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)

				// Stop here if its a preflight OPTIONS request
				if req.Method == "OPTIONS" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}
		// Dispatch to the next handler
		handler(w, req, p)
	}
}

func allowedOriginRegexps(allowedOrigins []string) []*regexp.Regexp {
	res, err := compileRegexps(allowedOrigins)
	if err != nil {
		klog.Fatalf("Invalid CORS allowed origin, --cors-allowed-origins flag was set to %v - %v", strings.Join(allowedOrigins, ","), err)
	}
	return res
}

// Takes a list of strings and compiles them into a list of regular expressions
func compileRegexps(regexpStrings []string) ([]*regexp.Regexp, error) {
	regexps := []*regexp.Regexp{}
	for _, regexpStr := range regexpStrings {
		r, err := regexp.Compile(regexpStr)
		if err != nil {
			return []*regexp.Regexp{}, err
		}
		regexps = append(regexps, r)
	}
	return regexps, nil
}
