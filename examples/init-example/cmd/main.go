package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/init-example/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	rootMux := http.NewServeMux()
	rootMux.Handle("/", api.NewRouter())

	srv := &http.Server{Addr: addr, Handler: rootMux}
	log.Printf("starting init-example on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
