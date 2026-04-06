package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/stdlib/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	srv := BuildServer(addr)
	log.Printf("starting stdlib example on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

// BuildServer constructs an *http.Server for the stdlib example. Separated
// out for testing so tests can inspect the server without starting it.
func BuildServer(addr string) *http.Server {
	rootMux := http.NewServeMux()
	rootMux.Handle("/", api.NewRouter())
	return &http.Server{Addr: addr, Handler: rootMux}
}
