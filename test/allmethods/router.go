package allmethods

import (
	"encoding/json"
	"net/http"
)

type AllMethodsModel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewRouter() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("GET /allmethods", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"method": "GET"})
	})

	r.HandleFunc("POST /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("PUT /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("PATCH /allmethods", func(w http.ResponseWriter, r *http.Request) {
		var in AllMethodsModel
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, in)
	})

	r.HandleFunc("DELETE /allmethods", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
