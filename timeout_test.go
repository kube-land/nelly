package nelly

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/julienschmidt/httprouter"

	"github.com/pharmatics/rest-util"
)

type recorder struct {
	lock  sync.Mutex
	count int
}

func (r *recorder) Record() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.count++
}

func (r *recorder) Count() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.count
}

func newHandler(responseCh <-chan string, panicCh <-chan interface{}, writeErrCh chan<- error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		select {
		case resp := <-responseCh:
			_, err := w.Write([]byte(resp))
			writeErrCh <- err
		case panicReason := <-panicCh:
			panic(panicReason)
		}
	}
}

func TestTimeout(t *testing.T) {
	origReallyCrash := runtime.ReallyCrash
	runtime.ReallyCrash = false
	defer func() {
		runtime.ReallyCrash = origReallyCrash
	}()

	sendResponse := make(chan string, 1)
	doPanic := make(chan interface{}, 1)
	writeErrors := make(chan error, 1)
	gotPanic := make(chan interface{}, 1)
	timeout := make(chan time.Time, 1)
	resp := "test response"
	timeoutErr := restutil.Error("request did not complete", restutil.StatusReasonTimeout)
	record := &recorder{}

	handler := newHandler(sendResponse, doPanic, writeErrors)

	router := httprouter.New()
	router.GET("/", withPanicRecovery(
		withTimeout(handler, func(req *http.Request) (*http.Request, <-chan time.Time, func(), *restutil.StatusError) {
			return req, timeout, record.Record, timeoutErr
		}), func(w http.ResponseWriter, req *http.Request, err interface{}) {
			gotPanic <- err
			http.Error(w, "This request caused panic. Look in the logs for details.", http.StatusInternalServerError)
		}),
	)

	ts := httptest.NewServer(router)
	defer ts.Close()

	// No timeouts
	sendResponse <- resp
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("got res.StatusCode %d; expected %d", res.StatusCode, http.StatusOK)
	}
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != resp {
		t.Errorf("got body %q; expected %q", string(body), resp)
	}
	if err := <-writeErrors; err != nil {
		t.Errorf("got unexpected Write error on first request: %v", err)
	}
	if record.Count() != 0 {
		t.Errorf("invoked record method: %#v", record)
	}

	// Times out
	timeout <- time.Time{}
	res, err = http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("got res.StatusCode %d; expected %d", res.StatusCode, http.StatusServiceUnavailable)
	}
	if record.Count() != 1 {
		t.Errorf("did not invoke record method: %#v", record)
	}

	// Now try to send a response
	sendResponse <- resp
	if err := <-writeErrors; err != http.ErrHandlerTimeout {
		t.Errorf("got Write error of %v; expected %v", err, http.ErrHandlerTimeout)
	}

	// Panics
	doPanic <- "inner handler panics"
	res, err = http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("got res.StatusCode %d; expected %d due to panic", res.StatusCode, http.StatusInternalServerError)
	}
	select {
	case err := <-gotPanic:
		msg := fmt.Sprintf("%v", err)
		if !strings.Contains(msg, "newHandler") {
			t.Errorf("expected line with root cause panic in the stack trace, but didn't: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("expected to see a handler panic, but didn't")
	}

	// Panics with http.ErrAbortHandler
	doPanic <- http.ErrAbortHandler
	res, err = http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("got res.StatusCode %d; expected %d due to panic", res.StatusCode, http.StatusInternalServerError)
	}
	select {
	case err := <-gotPanic:
		if err != http.ErrAbortHandler {
			t.Errorf("expected unwrapped http.ErrAbortHandler, got %#v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("expected to see a handler panic, but didn't")
	}
}
