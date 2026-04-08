package nestedpaths

import (
	"net/http"
	"strings"

	"github.com/oaswrap/gswag/test/util"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Order struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Total  int    `json:"total"`
}

type OrderItem struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	Product string `json:"product"`
	Qty     int    `json:"qty"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, []User{{ID: "u1", Name: "Alice"}})
	})

	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		util.WriteJSON(w, http.StatusOK, User{ID: id, Name: "Alice"})
	})

	mux.HandleFunc("GET /api/v1/users/{id}/orders", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		util.WriteJSON(w, http.StatusOK, []Order{{ID: "o1", UserID: id, Total: 99}})
	})

	mux.HandleFunc("GET /api/v1/users/{id}/orders/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		oid := r.PathValue("orderId")
		util.WriteJSON(w, http.StatusOK, Order{ID: oid, UserID: id, Total: 99})
	})

	mux.HandleFunc("GET /api/v1/users/{id}/orders/{orderId}/items", func(w http.ResponseWriter, r *http.Request) {
		oid := r.PathValue("orderId")
		util.WriteJSON(w, http.StatusOK, []OrderItem{{
			ID: "i1", OrderID: oid, Product: "widget", Qty: 2,
		}})
	})

	// Verify path concatenation produces the right URL by echoing the path.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		util.WriteJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path})
	})

	return mux
}
