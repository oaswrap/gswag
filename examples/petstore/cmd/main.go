package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/petstore/api"
	"github.com/oaswrap/gswag/examples/petstore/docs"
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/stoplight"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	router := api.NewRouter()

	// Stoplight Elements
	handler := specui.NewHandler(
		specui.WithTitle("Swagger Petstore"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)

	mux := http.NewServeMux()
	mux.Handle(handler.DocsPath(), handler.DocsFunc())
	mux.Handle(handler.SpecPath(), handler.SpecFunc())
	mux.Handle("/", router)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("starting petstore example on http://localhost%s", addr)
	log.Fatal(server.ListenAndServe())
}
