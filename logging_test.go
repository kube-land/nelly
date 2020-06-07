package nelly

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestDefaultStacktracePred(t *testing.T) {
	for _, x := range []int{101, 200, 204, 302, 400, 404} {
		if defaultStacktracePred(x) {
			t.Fatalf("should not log on %v by default", x)
		}
	}

	for _, x := range []int{500, 100} {
		if !defaultStacktracePred(x) {
			t.Fatalf("should log on %v by default", x)
		}
	}
}

func TestStatusIsNot(t *testing.T) {
	statusTestTable := []struct {
		status   int
		statuses []int
		want     bool
	}{
		{http.StatusOK, []int{}, true},
		{http.StatusOK, []int{http.StatusOK}, false},
		{http.StatusCreated, []int{http.StatusOK, http.StatusAccepted}, true},
	}
	for _, tt := range statusTestTable {
		sp := StatusIsNot(tt.statuses...)
		got := sp(tt.status)
		if got != tt.want {
			t.Errorf("Expected %v, got %v", tt.want, got)
		}
	}
}

func TestWithLogging(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	var handler httprouter.Handle
	handler = func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {}
	handler = withLogging(withLogging(handler, defaultStacktracePred), defaultStacktracePred)

	router := httprouter.New()
	router.GET("/v1", handler)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected newLogged to panic")
			}
		}()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}()
}

type testResponseWriter struct{}

func (*testResponseWriter) Header() http.Header       { return nil }
func (*testResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (*testResponseWriter) WriteHeader(int)           {}

func TestLoggedStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var tw http.ResponseWriter = new(testResponseWriter)
	logger := newLogged(req, tw)
	logger.Write(nil)

	if logger.status != http.StatusOK {
		t.Errorf("expected status after write to be %v, got %v", http.StatusOK, logger.status)
	}

	tw = new(testResponseWriter)
	logger = newLogged(req, tw)
	logger.WriteHeader(http.StatusForbidden)
	logger.Write(nil)

	if logger.status != http.StatusForbidden {
		t.Errorf("expected status after write to remain %v, got %v", http.StatusForbidden, logger.status)
	}
}
