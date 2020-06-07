package nelly

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestCORSAllowedOrigins(t *testing.T) {
	table := []struct {
		allowedOrigins []string
		origin         string
		allowed        bool
	}{
		{[]string{}, "example.com", false},
		{[]string{"example.com"}, "example.com", true},
		{[]string{"example.com"}, "not-allowed.com", false},
		{[]string{"not-matching.com", "example.com"}, "example.com", true},
		{[]string{".*"}, "example.com", true},
	}

	for _, item := range table {
		handler := withCORS(
			func(http.ResponseWriter, *http.Request, httprouter.Params) {},
			item.allowedOrigins, nil, nil, nil, "true",
		)
		var response *http.Response
		func() {
			router := httprouter.New()
			router.GET("/version", handler)

			server := httptest.NewServer(router)
			defer server.Close()

			request, err := http.NewRequest("GET", server.URL+"/version", nil)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			request.Header.Set("Origin", item.origin)
			client := http.Client{}
			response, err = client.Do(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		if item.allowed {
			if !reflect.DeepEqual(item.origin, response.Header.Get("Access-Control-Allow-Origin")) {
				t.Errorf("Expected %#v, Got %#v", item.origin, response.Header.Get("Access-Control-Allow-Origin"))
			}

			if response.Header.Get("Access-Control-Allow-Credentials") == "" {
				t.Errorf("Expected Access-Control-Allow-Credentials header to be set")
			}

			if response.Header.Get("Access-Control-Allow-Headers") == "" {
				t.Errorf("Expected Access-Control-Allow-Headers header to be set")
			}

			if response.Header.Get("Access-Control-Allow-Methods") == "" {
				t.Errorf("Expected Access-Control-Allow-Methods header to be set")
			}

			if response.Header.Get("Access-Control-Expose-Headers") != "Date" {
				t.Errorf("Expected Date in Access-Control-Expose-Headers header")
			}
		} else {
			if response.Header.Get("Access-Control-Allow-Origin") != "" {
				t.Errorf("Expected Access-Control-Allow-Origin header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Credentials") != "" {
				t.Errorf("Expected Access-Control-Allow-Credentials header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Headers") != "" {
				t.Errorf("Expected Access-Control-Allow-Headers header to not be set")
			}

			if response.Header.Get("Access-Control-Allow-Methods") != "" {
				t.Errorf("Expected Access-Control-Allow-Methods header to not be set")
			}

			if response.Header.Get("Access-Control-Expose-Headers") == "Date" {
				t.Errorf("Expected Date in Access-Control-Expose-Headers header")
			}
		}
	}
}

func TestCORSAllowedMethods(t *testing.T) {
	tests := []struct {
		allowedMethods []string
		method         string
		allowed        bool
	}{
		{nil, "POST", true},
		{nil, "GET", true},
		{nil, "OPTIONS", true},
		{nil, "PUT", true},
		{nil, "DELETE", true},
		{nil, "PATCH", true},
		{[]string{"GET", "POST"}, "PATCH", false},
	}

	allowsMethod := func(res *http.Response, method string) bool {
		allowedMethods := strings.Split(res.Header.Get("Access-Control-Allow-Methods"), ",")
		for _, allowedMethod := range allowedMethods {
			if strings.TrimSpace(allowedMethod) == method {
				return true
			}
		}
		return false
	}

	for _, test := range tests {
		handler := withCORS(
			func(http.ResponseWriter, *http.Request, httprouter.Params) {},
			[]string{".*"}, test.allowedMethods, nil, nil, "true",
		)
		var response *http.Response
		func() {
			router := httprouter.New()
			router.Handle(test.method, "/version", handler)

			server := httptest.NewServer(router)
			defer server.Close()

			request, err := http.NewRequest(test.method, server.URL+"/version", nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			request.Header.Set("Origin", "allowed.com")
			client := http.Client{}
			response, err = client.Do(request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}()

		methodAllowed := allowsMethod(response, test.method)
		switch {
		case test.allowed && !methodAllowed:
			t.Errorf("Expected %v to be allowed, Got only %#v", test.method, response.Header.Get("Access-Control-Allow-Methods"))
		case !test.allowed && methodAllowed:
			t.Errorf("Unexpected allowed method %v, Expected only %#v", test.method, response.Header.Get("Access-Control-Allow-Methods"))
		}
	}

}

func TestCompileRegex(t *testing.T) {
	uncompiledRegexes := []string{"endsWithMe$", "^startingWithMe"}
	regexes, err := compileRegexps(uncompiledRegexes)

	if err != nil {
		t.Errorf("Failed to compile legal regexes: '%v': %v", uncompiledRegexes, err)
	}
	if len(regexes) != len(uncompiledRegexes) {
		t.Errorf("Wrong number of regexes returned: '%v': %v", uncompiledRegexes, regexes)
	}

	if !regexes[0].MatchString("Something that endsWithMe") {
		t.Errorf("Wrong regex returned: '%v': %v", uncompiledRegexes[0], regexes[0])
	}
	if regexes[0].MatchString("Something that doesn't endsWithMe.") {
		t.Errorf("Wrong regex returned: '%v': %v", uncompiledRegexes[0], regexes[0])
	}
	if !regexes[1].MatchString("startingWithMe is very important") {
		t.Errorf("Wrong regex returned: '%v': %v", uncompiledRegexes[1], regexes[1])
	}
	if regexes[1].MatchString("not startingWithMe should fail") {
		t.Errorf("Wrong regex returned: '%v': %v", uncompiledRegexes[1], regexes[1])
	}
}
