package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Put("/pet", func(w http.ResponseWriter, r *http.Request) {
		var p Pet
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, p)
	})

	r.Post("/pet", func(w http.ResponseWriter, r *http.Request) {
		var p Pet
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if p.ID == 0 {
			p.ID = 2
		}
		writeJSON(w, http.StatusOK, p)
	})

	r.Get("/pet/findByStatus", func(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusOK, res)
	})

	r.Get("/pet/findByTags", func(w http.ResponseWriter, r *http.Request) {
		_ = r.URL.Query().Get("tags")
		writeJSON(w, http.StatusOK, pets)
	})

	r.Get("/pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
		for _, p := range pets {
			if p.ID == id {
				writeJSON(w, http.StatusOK, p)
				return
			}
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	r.Post("/pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "petId"), 10, 64)
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
		writeJSON(w, http.StatusOK, pet)
	})

	r.Delete("/pet/{petId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Post("/pet/{petId}/uploadImage", func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{Code: 200, Type: "success", Message: "uploaded"}
		writeJSON(w, http.StatusOK, resp)
	})

	r.Get("/store/inventory", func(w http.ResponseWriter, r *http.Request) {
		inv := map[string]int{"available": 1, "pending": 0, "sold": 0}
		writeJSON(w, http.StatusOK, inv)
	})

	r.Post("/store/order", func(w http.ResponseWriter, r *http.Request) {
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if o.ID == 0 {
			o.ID = 2
		}
		writeJSON(w, http.StatusOK, o)
	})

	r.Get("/store/order/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "orderId"), 10, 64)
		for _, o := range orders {
			if o.ID == id {
				writeJSON(w, http.StatusOK, o)
				return
			}
		}
		fallback := orders[0]
		fallback.ID = id
		writeJSON(w, http.StatusOK, fallback)
	})

	r.Delete("/store/order/{orderId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Post("/user", func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		if u.ID == 0 {
			u.ID = 2
		}
		writeJSON(w, http.StatusOK, u)
	})

	r.Post("/user/createWithList", func(w http.ResponseWriter, r *http.Request) {
		var in []User
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, in)
	})

	r.Get("/user/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit", "500")
		w.Header().Set("X-Expires-After", "2026-01-01T00:00:00Z")
		writeJSON(w, http.StatusOK, map[string]string{"token": "ok"})
	})

	r.Get("/user/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/user/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		for _, u := range users {
			if strings.EqualFold(u.Username, username) {
				writeJSON(w, http.StatusOK, u)
				return
			}
		}
		fallback := users[0]
		fallback.Username = username
		writeJSON(w, http.StatusOK, fallback)
	})

	r.Put("/user/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		var in User
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}
		if in.ID == 0 {
			in.ID = 1
		}
		in.Username = username
		writeJSON(w, http.StatusOK, in)
	})

	r.Delete("/user/{username}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
