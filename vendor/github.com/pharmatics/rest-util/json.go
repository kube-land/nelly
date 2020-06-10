package restutil

import (
	"encoding/json"
	"net/http"
)

// ResponseJSON convert struct to JSON and writes JSON response with the specified status code.
func ResponseJSON(response interface{}, w http.ResponseWriter, statusCode int) {
	if response != nil {
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(jsonResponse)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	}
}
