package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/echo/api"
	"github.com/oaswrap/gswag/examples/echo/docs"
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/stoplight"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	app := api.NewRouter()

	// Stoplight Elements
	handler := specui.NewHandler(
		specui.WithTitle("Products API (Echo)"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)

	mux := http.NewServeMux()
	mux.Handle(handler.DocsPath(), handler.DocsFunc())
	mux.Handle(handler.SpecPath(), handler.SpecFunc())
	mux.Handle("/", app)

	server := &http.Server{Addr: addr, Handler: mux}

	log.Printf("starting echo example on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
