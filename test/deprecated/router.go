package deprecated

import (
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

type LegacyItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ItemV2 struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// v1 — deprecated endpoint.
	mux.HandleFunc("GET /v1/items", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, []LegacyItem{
			{ID: 1, Name: "Widget"},
			{ID: 2, Name: "Gadget"},
		})
	})

	mux.HandleFunc("GET /v1/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, LegacyItem{ID: 1, Name: "Widget"})
	})

	// v2 — current endpoint.
	mux.HandleFunc("GET /v2/items", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, []ItemV2{
			{ID: 1, DisplayName: "Widget Pro", Slug: "widget-pro"},
		})
	})

	return mux
}
