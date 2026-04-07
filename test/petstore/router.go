package petstore

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/oaswrap/gswag/test/util"
)

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Category  Category `json:"category"`
	PhotoURLs []string `json:"photoUrls"`
	Tags      []Tag    `json:"tags"`
	Status    string   `json:"status"`
}

type Order struct {
	ID       int64  `json:"id"`
	PetID    int64  `json:"petId"`
	Quantity int    `json:"quantity"`
	ShipDate string `json:"shipDate"`
	Status   string `json:"status"`
	Complete bool   `json:"complete"`
}

type User struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Phone      string `json:"phone"`
	UserStatus int    `json:"userStatus"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

var pets = []Pet{
	{
		ID:        1,
		Name:      "doggie",
		Category:  Category{ID: 1, Name: "Dogs"},
		PhotoURLs: []string{"https://example.com/dog.jpg"},
		Tags:      []Tag{{ID: 1, Name: "friendly"}},
		Status:    "available",
	},
}

var orders = []Order{
	{ID: 1, PetID: 1, Quantity: 1, ShipDate: "2026-01-01T00:00:00Z", Status: "placed", Complete: false},
}

var users = []User{
	{ID: 1, Username: "user1", FirstName: "John", LastName: "James", Email: "john@email.com", Password: "12345", Phone: "12345", UserStatus: 1},
}

func NewRouter() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("PUT /pet", func(w http.ResponseWriter, r *http.Request) {
		var p Pet
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		util.WriteJSON(w, http.StatusOK, p)
	})

	r.HandleFunc("POST /pet", func(w http.ResponseWriter, r *http.Request) {
		var p Pet
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if p.ID == 0 {
			p.ID = 2
		}
		util.WriteJSON(w, http.StatusOK, p)
	})

	r.HandleFunc("GET /pet/findByStatus", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "available"
		}
		res := make([]Pet, 0)
		for _, p := range pets {
			if p.Status == status {
				res = append(res, p)
			}
		}
		util.WriteJSON(w, http.StatusOK, res)
	})

	r.HandleFunc("GET /pet/findByTags", func(w http.ResponseWriter, r *http.Request) {
		_ = r.URL.Query().Get("tags")
		util.WriteJSON(w, http.StatusOK, pets)
	})

	r.HandleFunc("GET /pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("petId"), 10, 64)
		for _, p := range pets {
			if p.ID == id {
				util.WriteJSON(w, http.StatusOK, p)
				return
			}
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	r.HandleFunc("POST /pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("petId"), 10, 64)
		name := r.URL.Query().Get("name")
		status := r.URL.Query().Get("status")
		pet := pets[0]
		pet.ID = id
		if name != "" {
			pet.Name = name
		}
		if status != "" {
			pet.Status = status
		}
		util.WriteJSON(w, http.StatusOK, pet)
	})

	r.HandleFunc("DELETE /pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("POST /pet/{petId}/uploadImage", func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{Code: 200, Type: "success", Message: "uploaded"}
		util.WriteJSON(w, http.StatusOK, resp)
	})

	r.HandleFunc("GET /store/inventory", func(w http.ResponseWriter, r *http.Request) {
		inv := map[string]int{"available": 1, "pending": 0, "sold": 0}
		util.WriteJSON(w, http.StatusOK, inv)
	})

	r.HandleFunc("POST /store/order", func(w http.ResponseWriter, r *http.Request) {
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if o.ID == 0 {
			o.ID = 2
		}
		util.WriteJSON(w, http.StatusOK, o)
	})

	r.HandleFunc("GET /store/order/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("orderId"), 10, 64)
		for _, o := range orders {
			if o.ID == id {
				util.WriteJSON(w, http.StatusOK, o)
				return
			}
		}
		fallback := orders[0]
		fallback.ID = id
		util.WriteJSON(w, http.StatusOK, fallback)
	})

	r.HandleFunc("DELETE /store/order/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("POST /user", func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if u.ID == 0 {
			u.ID = 2
		}
		util.WriteJSON(w, http.StatusOK, u)
	})

	r.HandleFunc("POST /user/createWithList", func(w http.ResponseWriter, r *http.Request) {
		var in []User
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		util.WriteJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("GET /user/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit", "500")
		w.Header().Set("X-Expires-After", "2026-01-01T00:00:00Z")
		util.WriteJSON(w, http.StatusOK, map[string]string{"token": "ok"})
	})

	r.HandleFunc("GET /user/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.HandleFunc("GET /user/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		for _, u := range users {
			if strings.EqualFold(u.Username, username) {
				util.WriteJSON(w, http.StatusOK, u)
				return
			}
		}
		fallback := users[0]
		fallback.Username = username
		util.WriteJSON(w, http.StatusOK, fallback)
	})

	r.HandleFunc("PUT /user/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		var in User
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		if in.ID == 0 {
			in.ID = 1
		}
		in.Username = username
		util.WriteJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("DELETE /user/{username}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
