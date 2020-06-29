# rest-util
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/pharmatics/rest-util)
[![Go Report Card](https://goreportcard.com/badge/github.com/pharmatics/rest-util)](https://goreportcard.com/report/github.com/pharmatics/rest-util)
[![Coverage](http://gocover.io/_badge/github.com/pharmatics/rest-util)](http://gocover.io/github.com/pharmatics/rest-util)

The rest-util is a small library for handling RESTful JSON response and status messages. It could be used by any RESTful API easily to convert any structure to JSON response message and write to `ResponseWriter`.

## Usage

```go
package main

import (
    "net/http"
    "github.com/pharmatics/rest-util"
)

type User struct {
    Name string `json:"name,omitempty"`
    Age  int    `json:"age,omitempty"`
}

func main() {
    http.HandleFunc("/success", UserSuccess)
    http.HandleFunc("/failure", UserFailure)
    http.ListenAndServe(":8080", nil)
}

func UserSuccess(w http.ResponseWriter, r *http.Request) {
    user := User{
        Name: "Alice",
        Age: 23,
    }
    restutil.ResponseJSON(user, w, 200)
}

func UserFailure(w http.ResponseWriter, r *http.Request) {
    user := User{
        Name: "Alice",
        Age: 23,
    }
    status := restutil.ErrorWithDetails("Can't process user", restutil.StatusReasonInvalid, user)
    restutil.ResponseJSON(status, w, status.Code)
}
```