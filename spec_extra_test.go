package gswag

import (
	"net/http"
	"testing"
)

func TestNewSpecCollector_InitialisesSpecAndSecurity(t *testing.T) {
	cfg := &Config{Title: "T", Version: "v", Servers: []ServerConfig{{URL: "https://api.example"}}, SecuritySchemes: map[string]SecuritySchemeConfig{"k": BearerJWT()}}
	sc := newSpecCollector(cfg)
	if sc == nil || sc.reflector == nil || sc.reflector.Spec == nil {
		t.Fatalf("expected reflector.spec initialised")
	}
	if sc.reflector.Spec.Info.Title != "T" || sc.reflector.Spec.Info.Version != "v" {
		t.Fatalf("spec info not set")
	}
	if len(sc.reflector.Spec.Servers) == 0 {
		t.Fatalf("servers not set")
	}
	if sc.reflector.Spec.Components == nil || sc.reflector.Spec.Components.SecuritySchemes == nil {
		t.Fatalf("components.securityschemes not initialised")
	}
}

func TestRegister_InfersSchemaAndAppendsExamplesAndHeaders(t *testing.T) {
	// ensure global config enabled for examples
	prevCfg := globalConfig
	globalConfig = &Config{Title: "T", Version: "v", CaptureExamples: true, MaxExampleBytes: 1024}
	defer func() { globalConfig = prevCfg }()

	sc := newSpecCollector(globalConfig)

	b := newRequestBuilder("GET", "/items/{id}")
	b.summary = "list"
	b.tags = []string{"items"}
	b.pathParams["id"] = "1"
	// declare a typed request body (struct) so RequestBody example path is exercised
	b.body = struct{ Payload int }{Payload: 1}

	// declare response header schema to be attached
	b.respHeaders[200] = map[string]interface{}{"X-Count": 5}

	// recorded response with JSON body and request body bytes
	rec := &recordedResponse{
		StatusCode:       200,
		BodyBytes:        []byte(`{"id":"1","name":"bob"}`),
		Headers:          http.Header{"Content-Type": {"application/json"}},
		RequestBodyBytes: []byte(`{"payload":1}`),
		builder:          b,
	}

	sc.Register(b, rec)

	pi, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[b.path]
	if !ok {
		t.Fatalf("path not registered")
	}
	op, ok := pi.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("get operation not found")
	}

	// response schema should be inferred and attached
	ror, ok := op.Responses.MapOfResponseOrRefValues["200"]
	if !ok || ror.Response == nil || ror.Response.Content == nil {
		t.Fatalf("response content not present")
	}
	if _, found := ror.Response.Content["application/json"]; !found {
		t.Fatalf("application/json content not attached")
	}

	// response example should be attached
	if mt, found := ror.Response.Content["application/json"]; !found || mt.Example == nil {
		t.Fatalf("response example not attached")
	}

	// response header should have been appended
	if ror.Response.Headers == nil {
		t.Fatalf("response headers not present")
	}
	if _, ok := ror.Response.Headers["X-Count"]; !ok {
		t.Fatalf("expected X-Count header to be present")
	}
}
