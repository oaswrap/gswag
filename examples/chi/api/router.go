// Package api is a minimal Chi HTTP server used by the gswag Chi example tests.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Order is the domain model for the orders resource.
type Order struct {
	ID       string  `json:"id"`
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Total    float64 `json:"total"`
}

// CreateOrderRequest is the request body for placing an order.
type CreateOrderRequest struct {
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

var orders = []Order{
	{ID: "ord-1", Product: "Widget", Quantity: 2, Total: 19.98},
	{ID: "ord-2", Product: "Gadget", Quantity: 1, Total: 24.99},
}

// NewRouter returns a configured Chi router.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	})

	r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		order := Order{
			ID:       "ord-3",
			Product:  req.Product,
			Quantity: req.Quantity,
			Total:    float64(req.Quantity) * 9.99,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
	})

	r.Get("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		for _, o := range orders {
			if o.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(o)
				return
			}
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	r.Delete("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return r
}
