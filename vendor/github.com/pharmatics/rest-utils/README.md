# rest-utils
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](http://godoc.org/github.com/pharmatics/rest-utils)

The rest-utils is a small library for handling RESTful JSON response and status messages. It could be used by any RESTful API easily to convert any structure to JSON response message and write to `ResponseWriter`.

## Usage

```go
package main

import (
    "net/http"
    "github.com/pharmatics/rest-utils"
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
    restutils.ResponseJSON(user, w, 200)
}

func UserFailure(w http.ResponseWriter, r *http.Request) {
    user := User{
        Name: "Alice",
        Age: 23,
    }
    status := restutils.NewInvalid("Can't process user", user)
    restutils.ResponseJSON(status, w, http.StatusUnprocessableEntity)
}
```