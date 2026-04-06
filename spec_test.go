package gswag

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/swaggest/openapi-go/openapi3"
)

// -----------------------------------------------------------------------------
// DSL registrations used by Ginkgo tests in this file
// -----------------------------------------------------------------------------

var _ = Path("/spec-items", func() {
	Get("List spec items", func() {
		Tag("items")
		Security("bearerAuth")

		Response(200, "ok", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

var _ = Path("/spec-items/{id}", func() {
	Get("Get spec item", func() {
		Tag("items")
		Parameter("id", InPath, Integer)

		Response(200, "ok", func() {
			SetParam("id", "42")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

type specItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var _ = Path("/spec-typed", func() {
	Get("Get typed item", func() {
		Tag("typed")

		Response(200, "ok", func() {
			ResponseSchema(new(specItem))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

// -----------------------------------------------------------------------------
// Ginkgo tests (these rely on the root suite's BeforeSuite in root_suite_test.go)
// -----------------------------------------------------------------------------

var _ = Describe("SpecCollector", func() {
	It("has no error-level validation issues after operations are registered", func() {
		issues := ValidateSpec()
		for _, iss := range issues {
			if iss.Severity == "error" {
				Fail("unexpected spec error: " + iss.String())
			}
		}
	})

	It("can write spec to an alternate path and read it back", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "test.yaml")
		Expect(WriteSpecTo(outPath, YAML)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("Root Suite API"))
	})

	It("can write spec in JSON format", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "test.json")
		Expect(WriteSpecTo(outPath, JSON)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(`"Root Suite API"`))
	})
})

// -----------------------------------------------------------------------------
// Standard `testing` package unit tests merged from the spec-related files
// -----------------------------------------------------------------------------

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

func TestDslSchemaTypeToReflect(t *testing.T) {
	if dslSchemaTypeToReflect(Integer) != reflect.TypeOf(int64(0)) {
		t.Fatalf("Integer did not map to int64")
	}
	if dslSchemaTypeToReflect(Number) != reflect.TypeOf(float64(0)) {
		t.Fatalf("Number did not map to float64")
	}
	if dslSchemaTypeToReflect(Boolean) != reflect.TypeOf(false) {
		t.Fatalf("Boolean did not map to bool")
	}
	if dslSchemaTypeToReflect(String) != reflect.TypeOf("") {
		t.Fatalf("String did not map to string")
	}
}

func TestDslSchemaParamAndStringParamJSON(t *testing.T) {
	p := dslSchemaParam("limit", Integer, openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
	if reflect.ValueOf(p).IsZero() {
		t.Fatalf("expected non-zero parameter")
	}

	sp := stringParam("q", openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
	if reflect.ValueOf(sp).IsZero() {
		t.Fatalf("expected non-zero string param")
	}
}

func TestBuildPathParamsStructFromDSLAndInfer(t *testing.T) {
	params := []dslParam{{name: "id", location: InPath, typ: Integer}}
	v := buildPathParamsStructFromDSL("/users/{id}", params)
	if v == nil {
		t.Fatalf("expected non-nil struct")
	}
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		t.Fatalf("expected pointer to struct, got %v", rt.Kind())
	}
	f := rt.Elem().Field(0)
	if f.Type != reflect.TypeOf(int64(0)) {
		t.Fatalf("expected field type int64, got %v", f.Type)
	}

	// buildPathParamsStruct should infer int when concrete value looks numeric
	v2 := buildPathParamsStruct("/items/{itemId}", map[string]string{"itemId": "42"})
	if v2 == nil {
		t.Fatalf("expected non-nil struct from buildPathParamsStruct")
	}
	rt2 := reflect.TypeOf(v2)
	f2 := rt2.Elem().Field(0)
	if f2.Type != reflect.TypeOf(int64(0)) {
		t.Fatalf("expected inferred int64 field, got %v", f2.Type)
	}
}

func TestCopyDslOpAndRespExecDeepCopy(t *testing.T) {
	op := &dslOp{
		method: "GET",
		path:   "/z",
		tags:   []string{"a"},
		params: []dslParam{{name: "p", location: InQuery, typ: String}},
		responses: map[int]*dslRespSpec{
			200: {description: "ok", headers: map[string]interface{}{"X": "v"}},
		},
	}

	copy := copyDslOp(op)
	// mutate original
	op.tags[0] = "b"
	op.params[0].name = "q"
	op.responses[200].headers["X"] = "changed"

	if copy.tags[0] != "a" {
		t.Fatalf("tags were not copied deeply")
	}
	if copy.params[0].name != "p" {
		t.Fatalf("params were not copied deeply")
	}
	if copy.responses[200].headers["X"] != "v" {
		t.Fatalf("response headers were not copied deeply")
	}

	// resp exec
	re := &dslRespExec{status: 200}
	re.pathParams = map[string]string{"id": "1"}
	re.queryParams = map[string]string{"q": "v"}
	re.headers = map[string]string{"H": "v"}
	re.bodyRaw = []byte("x")
	rcopy := copyDslRespExec(re)
	re.pathParams["id"] = "2"
	re.bodyRaw[0] = 'y'
	if rcopy.pathParams["id"] != "1" {
		t.Fatalf("resp exec pathParams not deeply copied")
	}
	if string(rcopy.bodyRaw) != "x" {
		t.Fatalf("resp exec bodyRaw not deeply copied")
	}
}

func TestRegisterDSLOperation_AppendsDSLParamsAndResponseHeaders(t *testing.T) {
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)

	op := &dslOp{
		method:  "GET",
		path:    "/things/{id}",
		summary: "get thing",
		params: []dslParam{
			{name: "q", location: InQuery, typ: String},
			{name: "X-Req", location: InHeader, typ: String},
		},
		responses: map[int]*dslRespSpec{
			200: {description: "ok", headers: map[string]interface{}{"X-Count": 1}},
		},
	}

	sc.RegisterDSLOperation(op)

	pi, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/things/{id}"]
	if !ok {
		t.Fatalf("path not registered")
	}
	oper, ok := pi.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("get operation not found")
	}

	// check params appended (query + header)
	foundQ := false
	foundH := false
	for _, p := range oper.Parameters {
		if p.Parameter.Name == "q" {
			foundQ = true
		}
		if p.Parameter.Name == "X-Req" {
			foundH = true
		}
	}
	if !foundQ || !foundH {
		t.Fatalf("expected DSL params appended q=%v h=%v", foundQ, foundH)
	}

	// check response header present
	r := oper.Responses.MapOfResponseOrRefValues["200"].Response
	if r == nil || r.Headers == nil {
		t.Fatalf("response or headers missing")
	}
	if _, ok := r.Headers["X-Count"]; !ok {
		t.Fatalf("expected X-Count header in response")
	}
}

func TestTopOpTopRespExecPanics(t *testing.T) {
	// ensure stacks empty
	dslOpStack = nil
	dslRespExecStack = nil

	// topOp should panic
	did := false
	func() {
		defer func() {
			if recover() != nil {
				did = true
			}
		}()
		_ = topOp()
	}()
	if !did {
		t.Fatalf("expected topOp to panic when called outside an operation")
	}

	did = false
	func() {
		defer func() {
			if recover() != nil {
				did = true
			}
		}()
		_ = topRespExec()
	}()
	if !did {
		t.Fatalf("expected topRespExec to panic when called outside a Response")
	}
}

// Helper used by schema inference tests
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

// jsonContains is a lightweight helper to check if a marshaled JSON blob contains a key or value string.
func jsonContains(b []byte, s string) bool {
	if !json.Valid(b) {
		return false
	}
	return containsString(string(b), s)
}

func containsString(hay, needle string) bool {
	return len(hay) > 0 && (func() bool { return (len(needle) > 0 && (len(hay) >= len(needle))) })() && (reflect.DeepEqual(true, true)) && (func() bool {
		return len(needle) == 0 || (len(hay) >= len(needle) && (func() bool { return stringIndex(hay, needle) >= 0 })())
	})()
}

func stringIndex(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
