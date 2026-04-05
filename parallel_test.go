package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
)

func setupParallelNode(t *testing.T, title, path string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	gswag.Init(&gswag.Config{
		Title:   title,
		Version: "1.0.0",
	})
	gswag.GET(path).WithTag("test").WithSummary("Test endpoint").Do(srv)
}

func TestWritePartialSpec(t *testing.T) {
	dir := t.TempDir()
	setupParallelNode(t, "Parallel API", "/items")

	if err := gswag.WritePartialSpec(1, dir); err != nil {
		t.Fatalf("WritePartialSpec failed: %v", err)
	}

	partialPath := filepath.Join(dir, "node-1.json")
	data, err := os.ReadFile(partialPath)
	if err != nil {
		t.Fatalf("partial spec file not found: %v", err)
	}
	if !strings.Contains(string(data), "Parallel API") {
		t.Errorf("expected title in partial spec, got:\n%s", string(data))
	}
}

func TestMergeAndWriteSpec(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "merged.yaml")

	// Simulate two nodes writing partial specs.
	// Node 1: /items
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1}`)) //nolint:errcheck
	}))
	defer srv1.Close()
	gswag.GET("/items").WithTag("items").WithSummary("List items").Do(srv1)
	if err := gswag.WritePartialSpec(1, dir); err != nil {
		t.Fatalf("node 1 WritePartialSpec: %v", err)
	}

	// Node 2: /orders — need a fresh collector via Init again.
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":2}`)) //nolint:errcheck
	}))
	defer srv2.Close()
	gswag.GET("/orders").WithTag("orders").WithSummary("List orders").Do(srv2)
	if err := gswag.WritePartialSpec(2, dir); err != nil {
		t.Fatalf("node 2 WritePartialSpec: %v", err)
	}

	// Restore node 1's config for merge.
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	if err := gswag.MergeAndWriteSpec(2, dir); err != nil {
		t.Fatalf("MergeAndWriteSpec: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("merged spec not found: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "/items") {
		t.Errorf("expected /items in merged spec")
	}
	if !strings.Contains(content, "/orders") {
		t.Errorf("expected /orders in merged spec")
	}
}

func TestMergeAndWriteSpec_MissingPartial(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "API",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "out.yaml"),
	})
	// No partial files written — should fail.
	err := gswag.MergeAndWriteSpec(1, dir)
	if err == nil {
		t.Fatal("expected error for missing partial spec")
	}
}

func TestMergeAndWriteSpec_WithSchemasAndSecurity(t *testing.T) {
	type ItemResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type OrderResponse struct {
		OrderID string `json:"order_id"`
		Amount  int    `json:"amount"`
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "merged.yaml")

	// Node 1: /items with typed response and bearer auth.
	gswag.Init(&gswag.Config{
		Title:   "Schema Merge API",
		Version: "1.0.0",
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"bearerAuth": gswag.BearerJWT(),
		},
		OutputPath: outPath,
	})
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1,"name":"Widget"}`)) //nolint:errcheck
	}))
	defer srv1.Close()
	gswag.GET("/items").
		WithTag("items").
		WithSummary("List items").
		WithBearerAuth().
		ExpectResponseBody(ItemResponse{}).
		Do(srv1)
	if err := gswag.WritePartialSpec(1, dir); err != nil {
		t.Fatalf("node 1 WritePartialSpec: %v", err)
	}

	// Node 2: /orders with typed response and API key.
	gswag.Init(&gswag.Config{
		Title:   "Schema Merge API",
		Version: "1.0.0",
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"apiKey": gswag.APIKeyHeader("X-API-Key"),
		},
		OutputPath: outPath,
	})
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"order_id":"o1","amount":100}`)) //nolint:errcheck
	}))
	defer srv2.Close()
	gswag.GET("/orders").
		WithTag("orders").
		WithSummary("List orders").
		WithSecurity("apiKey").
		ExpectResponseBody(OrderResponse{}).
		Do(srv2)
	if err := gswag.WritePartialSpec(2, dir); err != nil {
		t.Fatalf("node 2 WritePartialSpec: %v", err)
	}

	// Restore a config for merge.
	gswag.Init(&gswag.Config{
		Title:      "Schema Merge API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	if err := gswag.MergeAndWriteSpec(2, dir); err != nil {
		t.Fatalf("MergeAndWriteSpec: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("merged spec not found: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "/items") {
		t.Errorf("expected /items in merged spec")
	}
	if !strings.Contains(content, "/orders") {
		t.Errorf("expected /orders in merged spec")
	}
}

func TestMergeAndWriteSpec_JSON(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "merged.json")

	gswag.Init(&gswag.Config{
		Title:        "JSON Merge API",
		Version:      "1.0.0",
		OutputPath:   outPath,
		OutputFormat: gswag.JSON,
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`)) //nolint:errcheck
	}))
	defer srv.Close()
	gswag.GET("/ping").WithTag("ping").WithSummary("Ping").Do(srv)
	if err := gswag.WritePartialSpec(1, dir); err != nil {
		t.Fatalf("WritePartialSpec: %v", err)
	}

	if err := gswag.MergeAndWriteSpec(1, dir); err != nil {
		t.Fatalf("MergeAndWriteSpec: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("merged spec not found: %v", err)
	}
	if !strings.Contains(string(data), `"JSON Merge API"`) {
		t.Errorf("expected title in JSON merged spec")
	}
}
