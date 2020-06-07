package nelly

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestCacheControl(t *testing.T) {
	tests := []struct {
		name string
		path string

		startingHeader string
		expectedHeader string
	}{
		{
			name:           "simple",
			path:           "/v1",
			expectedHeader: "no-cache, private",
		},
		{
			name:           "already-set",
			path:           "/v2",
			startingHeader: "nonsense",
			expectedHeader: "nonsense",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handle := func(http.ResponseWriter, *http.Request, httprouter.Params) {
				//do nothing
			}

			wrapped := WithCacheControl()(handle)

			router := httprouter.New()
			router.GET(test.path, wrapped)

			testRequest, err := http.NewRequest(http.MethodGet, test.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			if len(test.startingHeader) > 0 {
				w.Header().Set("Cache-Control", test.startingHeader)
			}

			router.ServeHTTP(w, testRequest)
			actual := w.Header().Get("Cache-Control")

			if actual != test.expectedHeader {
				t.Fatal(actual)
			}
		})
	}

}
