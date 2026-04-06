package stdlib_test

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

func TestCaptureExamplesWritten(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "openapi.yaml")

	gswag.Init(&gswag.Config{
		Title:           "Example Capture",
		Version:         "1.0.0",
		OutputPath:      out,
		CaptureExamples: true,
		MaxExampleBytes: 0,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"msg":"hello"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	// Send request with JSON body to be captured as example.
	reqBody := map[string]string{"name": "alice"}
	res := gswag.POST("/echo").
		WithRequestBody(reqBody).
		WithTag("echo").
		WithSummary("Echo").
		Do(srv)

	if res.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	// Write spec to disk and read it back.
	if err := gswag.WriteSpecTo(out, gswag.YAML); err != nil {
		t.Fatalf("WriteSpecTo failed: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		if err2 := json.Unmarshal(data, spec); err2 != nil {
			t.Fatalf("parse spec: yaml: %v; json: %v", err, err2)
		}
	}

	pi, ok := spec.Paths.MapOfPathItemValues["/echo"]
	if !ok {
		t.Fatalf("/echo path missing in spec")
	}
	op, ok := pi.MapOfOperationValues["post"]
	if !ok {
		t.Fatalf("POST operation missing for /echo")
	}

	// Check request body example
	if op.RequestBody == nil || op.RequestBody.RequestBody == nil {
		t.Fatalf("request body missing for /echo POST")
	}
	rb := op.RequestBody.RequestBody
	mtReq, ok := rb.Content["application/json"]
	if !ok {
		t.Fatalf("request content type application/json missing")
	}
	if mtReq.Example == nil {
		t.Fatalf("expected request example to be present")
	}

	// Check response example
	ror, ok := op.Responses.MapOfResponseOrRefValues["200"]
	if !ok || ror.Response == nil {
		t.Fatalf("200 response missing for /echo POST")
	}
	resp := ror.Response
	mtResp, ok := resp.Content["application/json"]
	if !ok {
		t.Fatalf("response content type application/json missing")
	}
	if mtResp.Example == nil {
		t.Fatalf("expected response example to be present")
	}
}
