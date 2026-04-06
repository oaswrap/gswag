package gswag_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/oaswrap/gswag"
	"github.com/swaggest/openapi-go/openapi3"
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

func TestSpecCollector_ResponseHeaders(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "openapi.yaml")
	gswag.Init(&gswag.Config{
		Title:      "Header Spec",
		Version:    "1.0.0",
		OutputPath: out,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Rate-Limit", "5")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))
	defer srv.Close()

	// Declare a response header schema for 200 and execute the request with same builder.
	gswag.GET("/ping").
		WithTag("ping").
		WithSummary("Ping").
		ExpectResponseHeader("X-Rate-Limit", "").
		Do(srv)

	// Write spec and read it back.
	if err := gswag.WriteSpec(); err != nil {
		t.Fatalf("WriteSpec failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		// try JSON
		if err2 := json.Unmarshal(data, spec); err2 != nil {
			t.Fatalf("parse spec: yaml: %v; json: %v", err, err2)
		}
	}

	// Locate header schema.
	pi, ok := spec.Paths.MapOfPathItemValues["/ping"]
	if !ok {
		t.Fatalf("/ping path missing in spec")
	}
	op, ok := pi.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("GET operation missing for /ping")
	}
	ror, ok := op.Responses.MapOfResponseOrRefValues["200"]
	if !ok || ror.Response == nil {
		t.Fatalf("200 response missing for /ping GET")
	}
	if ror.Response.Headers == nil {
		t.Fatalf("response headers missing for 200")
	}
	if _, ok := ror.Response.Headers["X-Rate-Limit"]; !ok {
		t.Fatalf("expected X-Rate-Limit header in response headers")
	}
}
