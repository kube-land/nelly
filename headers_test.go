package nelly

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestWithRequiredHeaders(t *testing.T) {

	withRequiredHeaders := WithRequiredHeaders([]string{"X-Test-Header-1", "X-Test-Header-2"})

	router := httprouter.New()
	router.GET("/v1", withRequiredHeaders(func(http.ResponseWriter, *http.Request, httprouter.Params) {}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := http.Client{}

	// without header
	reqWithoutHeader, err := http.NewRequest(http.MethodGet, ts.URL+"/v1", nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.Do(reqWithoutHeader)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be %v, got %v", http.StatusBadRequest, response.StatusCode)
	}

	// with header
	reqWithHeader, err := http.NewRequest(http.MethodGet, ts.URL+"/v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqWithHeader.Header.Set("X-Test-Header-1", "any-value")
	reqWithHeader.Header.Set("X-Test-Header-2", "any-value")

	response, err = client.Do(reqWithHeader)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status to be %v, got %v", http.StatusOK, response.StatusCode)
	}

}

func TestWithRequiredHeaderValues(t *testing.T) {

	withRequiredHeadersValues := WithRequiredHeaderValues(map[string]string{
		"X-Test-Header-1": "1",
		"X-Test-Header-2": "2",
	})

	router := httprouter.New()
	router.GET("/v1", withRequiredHeadersValues(func(http.ResponseWriter, *http.Request, httprouter.Params) {}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := http.Client{}

	// with invalid header
	reqInvalidHeader, err := http.NewRequest(http.MethodGet, ts.URL+"/v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqInvalidHeader.Header.Set("X-Test-Header-1", "any-value")
	reqInvalidHeader.Header.Set("X-Test-Header-2", "2")

	resp, err := client.Do(reqInvalidHeader)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be %v, got %v", http.StatusBadRequest, resp.StatusCode)
	}

	// with proper header
	reqHeader, err := http.NewRequest(http.MethodGet, ts.URL+"/v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqHeader.Header.Set("X-Test-Header-1", "1")
	reqHeader.Header.Set("X-Test-Header-2", "2")

	resp, err = client.Do(reqHeader)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status to be %v, got %v", http.StatusOK, resp.StatusCode)
	}

}
