package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oaswrap/gswag/examples/stdlib/api"
	"github.com/oaswrap/gswag/examples/stdlib/docs"
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/stoplight"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	srv := BuildServer(addr)
	log.Printf("starting stdlib example on http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}

// BuildServer constructs an *http.Server for the stdlib example. Separated
// out for testing so tests can inspect the server without starting it.
func BuildServer(addr string) *http.Server {
	// Stoplight Elements
	uiHandler := specui.NewHandler(
		specui.WithTitle("Users API"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)

	rootMux := http.NewServeMux()
	rootMux.Handle(uiHandler.DocsPath(), uiHandler.DocsFunc())
	rootMux.Handle(uiHandler.SpecPath(), uiHandler.SpecFunc())
	rootMux.Handle("/", api.NewRouter())
	return &http.Server{Addr: addr, Handler: rootMux}
}
