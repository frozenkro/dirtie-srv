package api

import (
	"fmt"
	"net/http"
)

func Init() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, wrld!")
  })
  
  fmt.Println("Starting web server on :8080")

  if err := http.ListenAndServe(":8080", nil); err != nil {
    fmt.Printf("Web server error: %v\n", err)
  }
}
