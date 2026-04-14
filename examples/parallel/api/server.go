// Package api is a minimal stdlib HTTP server used by the gswag parallel example tests.
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// User is the domain model for the users resource.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Post is the domain model for the posts resource.
type Post struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreatePostRequest is the request body for creating a post.
type CreatePostRequest struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

var users = []User{
	{ID: "1", Name: "Alice", Email: "alice@example.com"},
	{ID: "2", Name: "Bob", Email: "bob@example.com"},
}

var posts = []Post{
	{ID: "1", Title: "Hello World", Body: "My first post", UserID: "1"},
	{ID: "2", Title: "Go Testing", Body: "Testing with Ginkgo", UserID: "2"},
}

// simulateLatency adds a small delay to every handler so that parallel
// Ginkgo nodes each pick up roughly equal numbers of specs, making the
// per-node partial specs visibly distinct.
const simulateLatency = 20 * time.Millisecond

// NewRouter returns a configured http.ServeMux.
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Users endpoints.
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(simulateLatency)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(users)
		case http.MethodPost:
			var req CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
				return
			}
			u := User{ID: "3", Name: req.Name, Email: req.Email}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(u)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(simulateLatency)
		id := strings.TrimPrefix(r.URL.Path, "/api/users/")
		for _, u := range users {
			if u.ID == id {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(u)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	// Posts endpoints.
	mux.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(simulateLatency)
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(posts)
		case http.MethodPost:
			var req CreatePostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
				return
			}
			p := Post{ID: "3", Title: req.Title, Body: req.Body, UserID: req.UserID}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(p)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(simulateLatency)
		id := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		for _, p := range posts {
			if p.ID == id {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(p)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})

	return mux
}
