package querystruct

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/oaswrap/gswag/test/util"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Tag   string `json:"tag"`
}

// ProductQuery defines query parameters for the product list endpoint.
type ProductQuery struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Sort     string `query:"sort"`
	Tag      string `query:"tag"`
}

var products = []Product{
	{ID: 1, Name: "Widget", Price: 10, Tag: "hardware"},
	{ID: 2, Name: "Gadget", Price: 25, Tag: "electronics"},
	{ID: 3, Name: "Sprocket", Price: 5, Tag: "hardware"},
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(q.Get("page_size"))
		if pageSize < 1 {
			pageSize = 10
		}

		var result []Product
		tag := q.Get("tag")
		for _, p := range products {
			if tag == "" || p.Tag == tag {
				result = append(result, p)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Total-Count", fmt.Sprintf("%d", len(result)))
		w.Header().Set("X-Page", fmt.Sprintf("%d", page))
		w.Header().Set("X-Page-Size", fmt.Sprintf("%d", pageSize))
		w.WriteHeader(http.StatusOK)
		util.WriteJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("GET /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
			return
		}
		for _, p := range products {
			if p.ID == id {
				util.WriteJSON(w, http.StatusOK, p)
				return
			}
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	return mux
}
