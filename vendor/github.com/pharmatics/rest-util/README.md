# rest-util

The rest-util is a small library for handling RESTful JSON response and status messages. It could be used by any RESTful API easily to convert any structure to JSON response message.

## Usage

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/pharmatics/rest-util"
)

type User struct {
    Name string `json:"name,omitempty"`
    Age  int    `json:"age,omitempty"`
}

func main() {
    http.HandleFunc("/", HelloServer)
    http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {

    user := User{
        Name: "Alice",
        Role: 23,
    }

    restutil.ResponseJSON(user, w, 200)
}
```