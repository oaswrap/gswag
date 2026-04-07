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

	router := api.NewRouter()
	addr := ":" + port

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Printf("starting petstore example on http://localhost%s", addr)
	log.Fatal(server.ListenAndServe())
}
