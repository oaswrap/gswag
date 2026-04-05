package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/oaswrap/gswag"
)

func TestSpecCollector_RegisterSimpleGET(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Spec Test",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`[{"id":1}]`)) //nolint:errcheck
	}))
	defer srv.Close()

	resp := gswag.GET("/items").
		WithTag("items").
		WithSummary("List items").
		WithQueryParam("limit", "10").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error after Register: %s", iss)
		}
	}
}

func TestSpecCollector_RegisterWithPathParam(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Spec Test",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":42}`)) //nolint:errcheck
	}))
	defer srv.Close()

	resp := gswag.GET("/items/{id}").
		WithPathParam("id", "42").
		WithTag("items").
		WithSummary("Get item").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Validate that no spec errors were raised (path param must be declared).
	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error after path param register: %s", iss)
		}
	}
}

func TestSpecCollector_RegisterWithIntPathParam(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Spec Test",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1}`)) //nolint:errcheck
	}))
	defer srv.Close()

	resp := gswag.GET("/orders/{id}").
		WithPathParam("id", "123"). // numeric → int64 field type
		WithSummary("Get order").
		WithTag("orders").
		Do(srv)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSpecCollector_RegisterWithBearerAuth(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Secure Spec",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.GET("/secure").
		WithBearerAuth().
		WithTag("secure").
		WithSummary("Secure endpoint").
		Do(srv)

	// bearerAuth must be auto-registered — validate should not error.
	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error: %s", iss)
		}
	}
}

func TestSpecCollector_RegisterWithTypedResponseBody(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Typed Response",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1,"name":"Widget"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.GET("/items/{id}").
		WithPathParam("id", "1").
		WithTag("items").
		WithSummary("Get item").
		ExpectResponseBody(Item{}).
		Do(srv)

	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error: %s", iss)
		}
	}
}

func TestSpecCollector_RegisterWithQueryParamStruct(t *testing.T) {
	type Filter struct {
		Page  int    `query:"page"`
		Limit int    `query:"limit"`
		Sort  string `query:"sort"`
	}

	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Param Struct",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`[]`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.GET("/items").
		WithQueryParamStruct(&Filter{}).
		WithTag("items").
		WithSummary("List items").
		Do(srv)

	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error with query param struct: %s", iss)
		}
	}
}

func TestSpecCollector_RegisterDeprecated(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Deprecated API",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	resp := gswag.DELETE("/items/{id}").
		WithPathParam("id", "1").
		WithTag("items").
		WithSummary("Delete item").
		AsDeprecated().
		Do(srv)

	if resp.StatusCode != 204 {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestSpecCollector_MultipleOperations(t *testing.T) {
	dir := t.TempDir()
	gswag.Init(&gswag.Config{
		Title:      "Multi Op",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(200)
			w.Write([]byte(`[]`)) //nolint:errcheck
		case http.MethodPost:
			w.WriteHeader(201)
			w.Write([]byte(`{"id":1}`)) //nolint:errcheck
		}
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	type Item struct {
		Name string `json:"name"`
	}

	gswag.GET("/things").WithTag("things").WithSummary("List things").Do(srv)
	gswag.POST("/things").
		WithTag("things").
		WithSummary("Create thing").
		WithRequestBody(Item{Name: "X"}).
		Do(srv)

	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error: %s", iss)
		}
	}
}
