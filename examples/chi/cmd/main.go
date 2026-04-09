package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/chi/api"
	"github.com/oaswrap/gswag/examples/chi/docs"
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
		specui.WithTitle("Pet Store API"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)
	r.Get(handler.DocsPath(), handler.DocsFunc())
	r.Get(handler.SpecPath(), handler.SpecFunc())

	server := &http.Server{Addr: addr, Handler: r}

	log.Printf("starting chi example on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
