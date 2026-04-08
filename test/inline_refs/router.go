package inlinerefs

import (
	"encoding/json"
	"net/http"

	"github.com/oaswrap/gswag/test/util"
)

// Address is a nested struct — with InlineRefs enabled the reflector should
// inline the schema rather than creating a $ref to #/components/schemas/Address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type Contact struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Customer struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Address Address `json:"address"`
	Contact Contact `json:"contact"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /customers", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, []Customer{
			{
				ID:   1,
				Name: "Alice",
				Address: Address{Street: "123 Main St", City: "Springfield", Country: "US"},
				Contact: Contact{Email: "alice@example.com", Phone: "555-0100"},
			},
		})
	})

	mux.HandleFunc("POST /customers", func(w http.ResponseWriter, r *http.Request) {
		var c Customer
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			util.WriteErrorJSON(w, http.StatusBadRequest, "invalid input")
			return
		}
		c.ID = 2
		util.WriteJSON(w, http.StatusCreated, c)
	})

	return mux
}
