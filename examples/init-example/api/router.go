// Package api is a minimal stdlib HTTP server used by the gswag init-example tests.
package api

import (
	"net/http"
)

// NewRouter returns a configured http.ServeMux.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
