package gswag

import (
    "net/http"
    "testing"
)

func TestSpecCollectorRegister_InferSchemaAndExamplesAndParams(t *testing.T) {
    // Initialize global config and collector with example capture enabled.
    Init(&Config{Title: "T", Version: "v", CaptureExamples: true, MaxExampleBytes: 1024})

    // Build a request builder with query/header and no explicit response schema.
    b := newRequestBuilder(http.MethodGet, "/pets/{id}")
    b.tags = []string{"pets"}
    b.summary = "Get pet"
    b.pathParams["id"] = "123"
    b.queryParams["q"] = "1"
    b.headers["X-Test"] = "v"

    // Declare a response header schema for status 200
    b.respHeaders = map[int]map[string]interface{}{
        200: {"X-Rate": "10"},
    }

    // Create a recorded response with JSON body and request body bytes.
    res := &recordedResponse{
        StatusCode:       200,
        Headers:          http.Header{"Content-Type": {"application/json"}},
        BodyBytes:        []byte(`{"name":"rex"}`),
        RequestBodyBytes: []byte(`{"echo":true}`),
    }

    // Register and exercise the code paths.
    if globalCollector == nil {
        t.Fatalf("globalCollector nil after Init")
    }
    globalCollector.Register(b, res)

    // Check that the spec contains the operation and inferred schema/example.
    si := globalCollector.reflector.Spec.Paths.MapOfPathItemValues["/pets/{id}"]
    if si.MapOfOperationValues == nil {
        t.Fatalf("no operations for path")
    }
    op := si.MapOfOperationValues["get"]
    if len(op.Responses.MapOfResponseOrRefValues) == 0 {
        t.Fatalf("operation missing or has no responses")
    }
    if op.Summary == nil || *op.Summary != "Get pet" {
        t.Fatalf("unexpected summary: %v", op.Summary)
    }

    // Ensure parameters include our query and header
    foundQ := false
    foundH := false
    for _, p := range op.Parameters {
        if p.Parameter.Name == "q" {
            foundQ = true
        }
        if p.Parameter.Name == "X-Test" {
            foundH = true
        }
    }
    if !foundQ || !foundH {
        t.Fatalf("expected query and header params appended, got q=%v h=%v", foundQ, foundH)
    }

    // Ensure response for 200 has an example and a schema
    ror := op.Responses.MapOfResponseOrRefValues["200"]
    if ror.Response == nil {
        t.Fatalf("response for 200 missing")
    }
    mt, ok := ror.Response.Content["application/json"]
    if !ok {
        t.Fatalf("application/json content missing")
    }
    if mt.Example == nil {
        t.Fatalf("expected example to be set")
    }
    if mt.Schema == nil {
        t.Fatalf("expected inferred schema to be set")
    }
}
