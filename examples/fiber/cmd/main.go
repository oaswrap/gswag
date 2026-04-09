package main

import (
	"log"
	"os"

	"github.com/gofiber/adaptor/v2"
	"github.com/oaswrap/gswag/examples/fiber/api"
	"github.com/oaswrap/gswag/examples/fiber/docs"
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
		specui.WithTitle("Reviews API (Fiber)"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)
	app.Get(handler.DocsPath(), adaptor.HTTPHandler(handler.DocsFunc()))
	app.Get(handler.SpecPath(), adaptor.HTTPHandler(handler.SpecFunc()))

	log.Printf("starting fiber example on http://localhost:%s", port)
	log.Fatal(app.Listen(addr))
}
