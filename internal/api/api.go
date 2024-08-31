package api

import (
	"fmt"
  "log"
	"net/http"
  "os"
)

type Adapter func(http.Handler) http.Handler

func Init() {
  logger := log.New(os.Stdout, "server: ", log.Lshortfile)
  
  rootHandler := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "request sent to root /\n")
  })

  http.Handle("/", adapt(rootHandler, Notify(logger)))
  
  fmt.Println("Starting web server on :8080")

  if err := http.ListenAndServe(":8080", nil); err != nil {
    fmt.Printf("Web server error: %v\n", err)
  }
}

func adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

func Notify(logger *log.Logger) Adapter {
  return func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      logger.Println("before")
      defer logger.Println("after")
      h.ServeHTTP(w, r)  
    })
  }
}
