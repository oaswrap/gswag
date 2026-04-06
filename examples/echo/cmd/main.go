package main

import (
	"log"
	"os"

	"github.com/oaswrap/gswag/examples/echo/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	app := api.NewRouter()

	log.Fatal(app.Start(addr))
}
