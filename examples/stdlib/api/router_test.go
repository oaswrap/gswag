package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_ListUsers(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/users")
	if err != nil {
		t.Fatalf("GET /api/users failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var arr []User
	if err := json.Unmarshal(data, &arr); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(arr) < 1 {
		t.Fatalf("expected non-empty users list")
	}
}

func TestNewRouter_CreateUser_BadRequest(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/users", "application/json", bytes.NewBufferString("notjson"))
	if err != nil {
		t.Fatalf("POST /api/users failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad request, got %d", resp.StatusCode)
	}
}

func TestNewRouter_CreateUser_Success(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	u := User{Name: "Charlie", Email: "charlie@example.com"}
	b, _ := json.Marshal(u)
	resp, err := http.Post(srv.URL+"/api/users", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/users failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var got User
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if got.ID == "" {
		t.Fatalf("expected created user to have ID")
	}
}

func TestNewRouter_GetUserByID_NotFound(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/users/999")
	if err != nil {
		t.Fatalf("GET /api/users/999 failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
