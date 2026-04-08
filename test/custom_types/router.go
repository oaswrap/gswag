package customtypes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oaswrap/gswag/test/util"
)

// Status is a custom named string type — the reflector should render it as a plain string.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending"
)

// UserID is a custom named integer type.
type UserID int64

// Page is a generic paginated response wrapper.
type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	PageNum  int `json:"page"`
	PageSize int `json:"page_size"`
}

// Nullable[T] is a generic nullable wrapper that marshals to/from a plain JSON value
// (or null), rather than the struct encoding used by database/sql Null* types.
type Nullable[T any] struct {
	Value T
	Valid bool
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(data, &n.Value); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// NullOf is a convenience constructor for a valid Nullable value.
func NullOf[T any](v T) Nullable[T] {
	return Nullable[T]{Value: v, Valid: true}
}

// NullableFields uses the local Nullable[T] type so the JSON wire format is a
// plain scalar ("Alice", 30, …) rather than {"String":"Alice","Valid":true}.
type NullableFields struct {
	Name   Nullable[string]  `json:"name"`
	Age    Nullable[int64]   `json:"age"`
	Score  Nullable[float64] `json:"score"`
	Active Nullable[bool]    `json:"active"`
}

// Item demonstrates custom named types embedded in a struct.
type Item struct {
	ID     UserID `json:"id"`
	Status Status `json:"status"`
	Name   string `json:"name"`
}

// Event uses time.Time and json.RawMessage.
type Event struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	OccurredAt time.Time       `json:"occurred_at"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, Page[Item]{
			Items: []Item{
				{ID: 1, Status: StatusActive, Name: "Widget"},
				{ID: 2, Status: StatusInactive, Name: "Gadget"},
			},
			Total:    2,
			PageNum:  1,
			PageSize: 10,
		})
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			util.WriteErrorJSON(w, http.StatusBadRequest, "invalid input")
			return
		}
		if item.ID == 0 {
			item.ID = 99
		}
		util.WriteJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("GET /nullable", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, NullableFields{
			Name:   NullOf("Alice"),
			Age:    NullOf(int64(30)),
			Score:  NullOf(9.5),
			Active: NullOf(true),
		})
	})

	mux.HandleFunc("POST /nullable", func(w http.ResponseWriter, r *http.Request) {
		var nf NullableFields
		if err := json.NewDecoder(r.Body).Decode(&nf); err != nil {
			util.WriteErrorJSON(w, http.StatusBadRequest, "invalid input")
			return
		}
		util.WriteJSON(w, http.StatusOK, nf)
	})

	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSON(w, http.StatusOK, []Event{
			{
				ID:         1,
				Name:       "signup",
				OccurredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Metadata:   json.RawMessage(`{"plan":"free"}`),
			},
		})
	})

	return mux
}
