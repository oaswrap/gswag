package examples

import (
	"encoding/json"
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	items := []Item{
		{ID: 1, Name: "Widget", Price: 10},
		{ID: 2, Name: "Gadget", Price: 25},
	}

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, items)
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		item.ID = len(items) + 1
		items = append(items, item)
		util.WriteJSON(w, http.StatusCreated, item)
	})

	return mux
}
