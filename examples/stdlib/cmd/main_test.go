package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag/examples/stdlib/api"
)

func TestCmdServer_HandlerWiresAPI(t *testing.T) {
	rootMux := http.NewServeMux()
	rootMux.Handle("/", api.NewRouter())

	srv := httptest.NewServer(rootMux)
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
	var users []api.User
	if err := json.Unmarshal(data, &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("expected at least one user")
	}
}

func TestBuildServer_ReturnsServer(t *testing.T) {
	srv := BuildServer(":1234")
	if srv == nil {
		t.Fatalf("expected server, got nil")
	}
	if srv.Addr != ":1234" {
		t.Fatalf("unexpected addr: %s", srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatalf("expected handler set")
	}
}
