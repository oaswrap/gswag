package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/petstore/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/", api.NewRouter())

	addr := ":" + port
	log.Printf("starting petstore example on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
