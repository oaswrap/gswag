package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"github.com/oaswrap/gswag/examples/todo/api"
	"github.com/oaswrap/gswag/examples/todo/docs"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
	specui "github.com/oaswrap/spec-ui"
	"github.com/oaswrap/spec-ui/stoplight"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("sqlite", "./todos.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := repository.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	router := api.NewRouter(db)

	handler := specui.NewHandler(
		specui.WithTitle("Todo API"),
		specui.WithSpecEmbedFS("openapi.yaml", &docs.FS),
		stoplight.WithUI(),
	)

	mux := http.NewServeMux()
	mux.Handle(handler.DocsPath(), handler.DocsFunc())
	mux.Handle(handler.SpecPath(), handler.SpecFunc())
	mux.Handle("/", router)

	server := &http.Server{Addr: ":" + port, Handler: mux}
	log.Printf("starting todo API on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
