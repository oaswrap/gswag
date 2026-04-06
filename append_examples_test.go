package gswag_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
	"github.com/swaggest/openapi-go/openapi3"
)

func TestAppendExamples_NoCapture(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "openapi.yaml")

	gswag.Init(&gswag.Config{
		Title:      "NoCapture",
		Version:    "1.0.0",
		OutputPath: out,
		// CaptureExamples left false
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.POST("/nc").WithRequestBody(map[string]string{"a": "b"}).Do(srv)

	if err := gswag.WriteSpecTo(out, gswag.YAML); err != nil {
		t.Fatalf("WriteSpecTo failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	pi, ok := spec.Paths.MapOfPathItemValues["/nc"]
	if !ok {
		t.Fatalf("/nc path missing")
	}
	op, ok := pi.MapOfOperationValues["post"]
	if !ok {
		t.Fatalf("post op missing")
	}

	// Request example should not be present when CaptureExamples is false
	if op.RequestBody != nil && op.RequestBody.RequestBody != nil {
		if mt, ok := op.RequestBody.RequestBody.Content["application/json"]; ok {
			if mt.Example != nil {
				t.Fatalf("expected no request example, but found one")
			}
		}
	}
}

func TestAppendExamples_CapAndSanitizer(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "openapi.yaml")

	gswag.Init(&gswag.Config{
		Title:           "CapSanitize",
		Version:         "1.0.0",
		OutputPath:      out,
		CaptureExamples: true,
		MaxExampleBytes: 16,
		Sanitizer: func(b []byte) []byte {
			// return a small valid JSON example
			return []byte(`{"san":"ok"}`)
		},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"resp":"value"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	long := strings.Repeat("x", 200)
	gswag.POST("/cap").WithRequestBody(map[string]string{"long": long}).Do(srv)

	if err := gswag.WriteSpecTo(out, gswag.YAML); err != nil {
		t.Fatalf("WriteSpecTo failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	// Marshal to JSON then inspect generically to avoid pointer Example layout details
	jb, _ := json.Marshal(spec)
	var doc map[string]interface{}
	if err := json.Unmarshal(jb, &doc); err != nil {
		t.Fatalf("json unmarshal spec: %v", err)
	}

	paths := doc["paths"].(map[string]interface{})
	p := paths["/cap"].(map[string]interface{})
	post := p["post"].(map[string]interface{})
	rb := post["requestBody"].(map[string]interface{})
	content := rb["content"].(map[string]interface{})
	app := content["application/json"].(map[string]interface{})
	ex := app["example"]
	if ex == nil {
		t.Fatalf("expected sanitized request example")
	}
	// Expect sanitized JSON object with key 'san'
	m, ok := ex.(map[string]interface{})
	if !ok || m["san"] != "ok" {
		t.Fatalf("sanitizer output not present in example: %#v", ex)
	}
}

func TestAppendExamples_ResponseNotJSON(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "openapi.yaml")

	gswag.Init(&gswag.Config{
		Title:           "RespNotJSON",
		Version:         "1.0.0",
		OutputPath:      out,
		CaptureExamples: true,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("plain text response")) //nolint:errcheck
	}))
	defer srv.Close()

	gswag.POST("/text").WithRequestBody(map[string]string{"k": "v"}).Do(srv)

	if err := gswag.WriteSpecTo(out, gswag.YAML); err != nil {
		t.Fatalf("WriteSpecTo failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		t.Fatalf("unmarshal spec: %v", err)
	}

	pi, ok := spec.Paths.MapOfPathItemValues["/text"]
	if !ok {
		t.Fatalf("/text path missing")
	}
	op, ok := pi.MapOfOperationValues["post"]
	if !ok {
		t.Fatalf("post op missing")
	}

	// Response content is text/plain — our appendExamples prefers application/json,
	// so response example should not be set on application/json content.
	if ror, found := op.Responses.MapOfResponseOrRefValues["200"]; found && ror.Response != nil {
		if mt, ok := ror.Response.Content["application/json"]; ok {
			if mt.Example != nil {
				t.Fatalf("expected no application/json response example for text/plain response")
			}
		}
	}
}
