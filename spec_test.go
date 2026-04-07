package gswag

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestNewSpecCollector_AppliesTopLevelMetadata(t *testing.T) {
	cfg := &Config{
		Title:          "Meta API",
		Version:        "1.2.3",
		Description:    "desc",
		TermsOfService: "https://example.com/terms",
		Contact:        &ContactConfig{Name: "API Team", Email: "team@example.com", URL: "https://example.com/contact"},
		License:        &LicenseConfig{Name: "Apache 2.0", URL: "https://example.com/license"},
		ExternalDocs:   &ExternalDocsConfig{Description: "More", URL: "https://example.com/docs"},
		Tags: []TagConfig{{
			Name:        "pets",
			Description: "pets tag",
			ExternalDocs: &ExternalDocsConfig{
				Description: "pets docs",
				URL:         "https://example.com/pets",
			},
		}},
	}

	sc := newSpecCollector(cfg)
	if sc.reflector.Spec.Info.TermsOfService == nil || *sc.reflector.Spec.Info.TermsOfService != cfg.TermsOfService {
		t.Fatalf("termsOfService not set")
	}
	if sc.reflector.Spec.Info.Contact == nil || sc.reflector.Spec.Info.Contact.Email == nil || *sc.reflector.Spec.Info.Contact.Email != cfg.Contact.Email {
		t.Fatalf("contact not set")
	}
	if sc.reflector.Spec.Info.License == nil || sc.reflector.Spec.Info.License.Name != cfg.License.Name {
		t.Fatalf("license not set")
	}
	if sc.reflector.Spec.ExternalDocs == nil || sc.reflector.Spec.ExternalDocs.URL != cfg.ExternalDocs.URL {
		t.Fatalf("external docs not set")
	}
	if len(sc.reflector.Spec.Tags) != 1 || sc.reflector.Spec.Tags[0].Name != "pets" {
		t.Fatalf("tags metadata not set")
	}
}

func TestBuildSecuritySchemeOrRef_OAuth2Implicit(t *testing.T) {
	sor := buildSecuritySchemeOrRef(SecuritySchemeConfig{
		Type:             "oauth2",
		AuthorizationURL: "https://petstore3.swagger.io/oauth/authorize",
		Scopes: map[string]string{
			"write:pets": "modify pets in your account",
			"read:pets":  "read your pets",
		},
	})

	if sor.SecurityScheme == nil || sor.SecurityScheme.OAuth2SecurityScheme == nil {
		t.Fatalf("expected oauth2 security scheme")
	}
	flows := sor.SecurityScheme.OAuth2SecurityScheme.Flows
	if flows.Implicit == nil {
		t.Fatalf("expected implicit oauth flow")
	}
	if flows.Implicit.AuthorizationURL != "https://petstore3.swagger.io/oauth/authorize" {
		t.Fatalf("unexpected authorization url: %q", flows.Implicit.AuthorizationURL)
	}
	if got := flows.Implicit.Scopes["write:pets"]; got != "modify pets in your account" {
		t.Fatalf("unexpected write:pets scope description: %q", got)
	}
}

func TestRegisterDSLOperation_HiddenSkipsSpecRegistration(t *testing.T) {
	sc := newSpecCollector(&Config{Title: "T", Version: "v"})
	sc.RegisterDSLOperation(&dslOp{
		method:  http.MethodGet,
		path:    "/hidden",
		summary: "hidden",
		hidden:  true,
		responses: map[int]*dslRespSpec{
			200: {description: "ok"},
		},
	})

	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/hidden"]; ok {
		t.Fatalf("expected hidden operation to be excluded from spec")
	}
}

func TestRegister_ExcludePathsSkipsSpecRegistration(t *testing.T) {
	sc := newSpecCollector(&Config{Title: "T", Version: "v", ExcludePaths: []string{"/internal/*", "/admin/health"}})

	b := newRequestBuilder(http.MethodGet, "/internal/users")
	b.summary = "internal"
	rec := &recordedResponse{StatusCode: 200, Headers: http.Header{"Content-Type": {"application/json"}}, BodyBytes: []byte(`{"ok":true}`)}
	sc.Register(b, rec)

	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/internal/users"]; ok {
		t.Fatalf("expected excluded builder path to be skipped")
	}

	b2 := newRequestBuilder(http.MethodGet, "/public/users")
	b2.summary = "public"
	sc.Register(b2, rec)
	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/public/users"]; !ok {
		t.Fatalf("expected non-excluded builder path to be registered")
	}
}

func TestRegisterDSLOperation_ExcludePathsSkipsSpecRegistration(t *testing.T) {
	sc := newSpecCollector(&Config{Title: "T", Version: "v", ExcludePaths: []string{"/internal/*", "/admin/health"}})

	sc.RegisterDSLOperation(&dslOp{
		method:  http.MethodGet,
		path:    "/admin/health",
		summary: "health",
		responses: map[int]*dslRespSpec{
			200: {description: "ok"},
		},
	})

	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/admin/health"]; ok {
		t.Fatalf("expected excluded DSL path to be skipped")
	}

	sc.RegisterDSLOperation(&dslOp{
		method:  http.MethodGet,
		path:    "/public/health",
		summary: "health",
		responses: map[int]*dslRespSpec{
			200: {description: "ok"},
		},
	})

	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/public/health"]; !ok {
		t.Fatalf("expected non-excluded DSL path to be registered")
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

func TestDSLExecution_AppendsExamples(t *testing.T) {
	prevCfg := globalConfig
	prevCollector := globalCollector
	defer func() {
		globalConfig = prevCfg
		globalCollector = prevCollector
	}()

	globalConfig = &Config{Title: "T", Version: "v", CaptureExamples: true, MaxExampleBytes: 1024}
	globalCollector = newSpecCollector(globalConfig)

	type payload struct {
		Name string `json:"name"`
	}

	op := &dslOp{
		method:       http.MethodPost,
		path:         "/dsl-capture",
		summary:      "capture",
		tags:         []string{"capture"},
		reqBodyModel: new(payload),
		responses: map[int]*dslRespSpec{
			200: {description: "ok", bodyModel: new(payload)},
		},
	}
	globalCollector.RegisterDSLOperation(op)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"alice"}`))
	}))
	defer ts.Close()

	b := dslBuildRequest(http.MethodPost, "/dsl-capture", &dslRespExec{body: &payload{Name: "alice"}})
	recorded := b.do(ts)
	globalCollector.injectInferredRequestSchema(b, recorded)
	globalCollector.injectRecordedResponseSchema(http.MethodPost, "/dsl-capture", recorded)
	globalCollector.appendExamples(b, recorded)

	opItem := globalCollector.reflector.Spec.Paths.MapOfPathItemValues["/dsl-capture"].MapOfOperationValues["post"]
	if opItem.RequestBody == nil || opItem.RequestBody.RequestBody == nil {
		t.Fatalf("expected request body")
	}
	requestMT := opItem.RequestBody.RequestBody.Content["application/json"]
	if requestMT.Example == nil {
		t.Fatalf("expected request example to be set")
	}
	responseMT := opItem.Responses.MapOfResponseOrRefValues["200"].Response.Content["application/json"]
	if responseMT.Example == nil {
		t.Fatalf("expected response example to be set")
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
	p := dslSchemaParam(dslParam{name: "limit", typ: Integer}, openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
	if reflect.ValueOf(p).IsZero() {
		t.Fatalf("expected non-zero parameter")
	}

	sp := stringParam("q", openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
	if reflect.ValueOf(sp).IsZero() {
		t.Fatalf("expected non-zero string param")
	}
}

func TestDslSchemaParam_RequiredAndExplode(t *testing.T) {
	req := true
	explode := true
	p := dslSchemaParam(
		dslParam{name: "tags", typ: Array, required: &req, explode: &explode},
		openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}},
	)
	if p.Parameter == nil {
		t.Fatalf("expected parameter")
	}
	if p.Parameter.Required == nil || !*p.Parameter.Required {
		t.Fatalf("expected required=true")
	}
	if p.Parameter.Explode == nil || !*p.Parameter.Explode {
		t.Fatalf("expected explode=true")
	}
	if p.Parameter.Schema == nil || p.Parameter.Schema.Schema == nil || p.Parameter.Schema.Schema.Items == nil {
		t.Fatalf("expected array items schema")
	}
}

func TestDslSchemaParam_EnumAndDefault(t *testing.T) {
	p := dslSchemaParam(
		dslParam{
			name:     "status",
			typ:      String,
			enumVals: []interface{}{"available", "pending", "sold"},
			defVal:   "available",
			hasDef:   true,
		},
		openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}},
	)
	if p.Parameter == nil || p.Parameter.Schema == nil || p.Parameter.Schema.Schema == nil {
		t.Fatalf("expected parameter schema")
	}
	s := p.Parameter.Schema.Schema
	if len(s.Enum) != 3 {
		t.Fatalf("expected enum length 3, got %d", len(s.Enum))
	}
	if s.Default == nil {
		t.Fatalf("expected default value")
	}
	if dv, ok := (*s.Default).(string); !ok || dv != "available" {
		t.Fatalf("unexpected default value: %#v", *s.Default)
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
