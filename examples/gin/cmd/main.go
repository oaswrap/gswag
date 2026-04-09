package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/gin/api"
	"github.com/oaswrap/gswag/examples/gin/docs"
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/stoplight"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	r := api.NewRouter()

	// Stoplight Elements
	handler := specui.NewHandler(
		specui.WithTitle("Items API (Gin)"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)

	mux := http.NewServeMux()
	mux.Handle(handler.DocsPath(), handler.DocsFunc())
	mux.Handle(handler.SpecPath(), handler.SpecFunc())
	mux.Handle("/", r)

	server := &http.Server{Addr: addr, Handler: mux}

	log.Printf("starting gin example on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
