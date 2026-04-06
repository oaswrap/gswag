package gswag

import (
	"strconv"
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func makeOpWithResponse(status int) openapi3.Operation {
	op := openapi3.Operation{}
	op.Responses = openapi3.Responses{MapOfResponseOrRefValues: map[string]openapi3.ResponseOrRef{}}
	ror := openapi3.ResponseOrRef{}
	r := openapi3.Response{}
	r.Content = map[string]openapi3.MediaType{}
	ror.WithResponse(r)
	op.Responses.MapOfResponseOrRefValues[strconv.Itoa(status)] = ror
	return op
}

func TestInjectInferredSchema_Array(t *testing.T) {
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)

	// prepare path and operation slot
	path := "/arr"
	pi := openapi3.PathItem{MapOfOperationValues: map[string]openapi3.Operation{"get": makeOpWithResponse(200)}}
	if sc.reflector.Spec.Paths.MapOfPathItemValues == nil {
		sc.reflector.Spec.Paths.MapOfPathItemValues = map[string]openapi3.PathItem{}
	}
	sc.reflector.Spec.Paths.MapOfPathItemValues[path] = pi

	b := newRequestBuilder("GET", path)

	rec := &recordedResponse{StatusCode: 200, BodyBytes: []byte(`[{"id":1},{"id":2}]`)}

	sc.injectInferredSchema(b, rec)

	pi2, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[path]
	if !ok {
		t.Fatalf("path missing after injection")
	}
	op, ok := pi2.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("operation missing")
	}
	ror, ok := op.Responses.MapOfResponseOrRefValues["200"]
	if !ok || ror.Response == nil || ror.Response.Content == nil {
		t.Fatalf("response content not set")
	}
	if _, found := ror.Response.Content["application/json"]; !found {
		t.Fatalf("expected application/json content for array response")
	}
}

func TestInjectInferredSchema_NestedObject(t *testing.T) {
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)

	path := "/nested"
	pi := openapi3.PathItem{MapOfOperationValues: map[string]openapi3.Operation{"get": makeOpWithResponse(200)}}
	if sc.reflector.Spec.Paths.MapOfPathItemValues == nil {
		sc.reflector.Spec.Paths.MapOfPathItemValues = map[string]openapi3.PathItem{}
	}
	sc.reflector.Spec.Paths.MapOfPathItemValues[path] = pi

	b := newRequestBuilder("GET", path)
	rec := &recordedResponse{StatusCode: 200, BodyBytes: []byte(`{"user":{"id":1,"name":"x"},"tags":["a","b"]}`)}

	sc.injectInferredSchema(b, rec)

	pi2, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[path]
	if !ok {
		t.Fatalf("path missing after injection")
	}
	op, ok := pi2.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("operation missing")
	}
	ror, ok := op.Responses.MapOfResponseOrRefValues["200"]
	if !ok || ror.Response == nil || ror.Response.Content == nil {
		t.Fatalf("response content not set")
	}
	if _, found := ror.Response.Content["application/json"]; !found {
		t.Fatalf("expected application/json content for nested response")
	}
}
