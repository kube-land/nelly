package nelly

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testUser struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

func TestResponseJSON(t *testing.T) {

	recorder := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		user := testUser{
			Name: "Alice",
			Age:  22,
		}
		ResponseJSON(user, w, http.StatusOK)
	}

	req, err := http.NewRequest("GET", "http://localhost:3000/foobar", nil)
	if err != nil {
		t.Error(err)
	}

	handler(recorder, req)

	resp := recorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status response 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected 'application/json' Content-Type, got '%s'", resp.Header.Get("Content-Type"))
	}

	if string(body) != "{\"name\":\"Alice\",\"age\":22}" {
		t.Errorf("expected {\"name\":\"Alice\",\"age\":22}, got '%s'", string(body))
	}

}

func TestEmptyResponseJSON(t *testing.T) {

	recorder := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		ResponseJSON(nil, w, 0)
	}

	req, err := http.NewRequest("GET", "http://localhost:3000/foobar", nil)
	if err != nil {
		t.Error(err)
	}

	handler(recorder, req)

	resp := recorder.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status response 200, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected 'application/json' Content-Type, got '%s'", resp.Header.Get("Content-Type"))
	}

	if string(body) != "" {
		t.Errorf("expected empty response body, got '%s'", string(body))
	}

}

func TestCodeToString(t *testing.T) {
	noContent := codeToString(204)
	if noContent != "204" {
		t.Errorf("expected status '204', got '%v'", noContent)
	}

	notModified := codeToString(304)
	if notModified != "304" {
		t.Errorf("expected status '304', got '%v'", notModified)
	}

	unknown := codeToString(707)
	if unknown != "707" {
		t.Errorf("expected status '707', got '%v'", unknown)
	}
}
