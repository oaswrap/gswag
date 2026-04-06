package gswag

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- DSL / constant helpers ---

func TestParamLocationConstants(t *testing.T) {
	if PathParam != InPath {
		t.Error("PathParam should equal InPath")
	}
	if QueryParam != InQuery {
		t.Error("QueryParam should equal InQuery")
	}
	if HeaderParam != InHeader {
		t.Error("HeaderParam should equal InHeader")
	}
	if CookieParam != InCookie {
		t.Error("CookieParam should equal InCookie")
	}
}

func TestSchemaTypeConstants(t *testing.T) {
	cases := []struct {
		val  SchemaType
		want string
	}{
		{String, "string"},
		{Integer, "integer"},
		{Number, "number"},
		{Boolean, "boolean"},
		{Object, "object"},
		{Array, "array"},
	}
	for _, tc := range cases {
		if string(tc.val) != tc.want {
			t.Errorf("SchemaType %q: want %q, got %q", tc.val, tc.want, string(tc.val))
		}
	}
}

// --- SetTestServer ---

func TestSetTestServer_AcceptsServer(t *testing.T) {
	// SetTestServer should not panic for valid targets.
	SetTestServer("http://localhost:9999")
	// Restore to nil for other tests.
	SetTestServer(nil)
}

func TestSetTestServer_AllowsArbitraryTargetType(t *testing.T) {
	// SetTestServer stores the target as-is; validation happens when requests run.
	SetTestServer(12345)
	SetTestServer(nil)
}

// --- request builder: Do & related behavior ---

func TestRequestBuilder_Do_SuccessRegisters(t *testing.T) {
	// setup server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if _, err := w.Write([]byte(`{"id":1}`)); err != nil {
			t.Fatalf("failed to write response body: %v", err)
		}
	}))
	defer srv.Close()

	// prepare collector
	prevCollector := globalCollector
	prevConfig := globalConfig
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)
	globalCollector = sc
	globalConfig = cfg
	defer func() { globalCollector = prevCollector; globalConfig = prevConfig }()

	b := newRequestBuilder("GET", "/test")
	b.summary = "ok"

	rec := b.do(srv)
	if rec.StatusCode != 200 {
		t.Fatalf("expected 200 got %d", rec.StatusCode)
	}

	// spec should have registered path
	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/test"]; !ok {
		t.Fatalf("expected spec to contain /test")
	}
}

func TestRequestBuilder_Do_ValidationWarnAndFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if _, err := w.Write([]byte("notjson")); err != nil {
			t.Fatalf("failed to write response body: %v", err)
		}
	}))
	defer srv.Close()

	prevCollector := globalCollector
	prevConfig := globalConfig
	defer func() { globalCollector = prevCollector; globalConfig = prevConfig }()

	cfg := &Config{Title: "T", Version: "v", EnforceResponseValidation: true, ValidationMode: "warn"}
	sc := newSpecCollector(cfg)
	globalCollector = sc
	globalConfig = cfg

	b := newRequestBuilder("GET", "/val")
	// declare typed resp model to trigger typed-model validation path
	b.respBodies[200] = struct{ ID int }{}

	// warn mode should not panic
	_ = b.do(srv)

	// now fail mode should panic on validation issues
	cfg2 := &Config{Title: "T", Version: "v", EnforceResponseValidation: true, ValidationMode: "fail"}
	globalConfig = cfg2
	globalCollector = newSpecCollector(cfg2)

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				// recovered value might be string
				if s, ok := r.(string); ok {
					if strings.Contains(s, "response does not match declared schema") || strings.Contains(s, "response validation error") {
						didPanic = true
					}
				}
			}
		}()
		_ = b.do(srv)
	}()
	if !didPanic {
		t.Fatalf("expected panic on validation fail mode")
	}
}

func TestRequestBuilder_Do_NetworkFailurePanics(t *testing.T) {
	// server that we close before request to force connection error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	urlSrv := srv
	srv.Close()

	b := newRequestBuilder("GET", "/path")

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				if s, ok := r.(string); ok {
					if strings.Contains(s, "HTTP request failed") {
						didPanic = true
					}
				}
			}
		}()
		_ = b.do(urlSrv)
	}()
	if !didPanic {
		t.Fatalf("expected network failure to panic")
	}
}

// --- builder extra helpers ---

func TestResolveBaseURLAndBuildRequest(t *testing.T) {
	// string base URL
	s := resolveBaseURL("http://example.com/")
	if s != "http://example.com" {
		t.Fatalf("unexpected base url: %s", s)
	}

	// httptest server
	srv := httptest.NewServer(nil)
	defer srv.Close()
	s2 := resolveBaseURL(srv)
	if s2 != srv.URL {
		t.Fatalf("expected %s got %s", srv.URL, s2)
	}

	// buildRequest with body, headers, query
	b := newRequestBuilder("POST", "/path")
	b.body = map[string]interface{}{"a": 1}
	b.headers["X-Req"] = "v"
	b.queryParams["q"] = "1"

	req, data, err := b.buildRequest("http://example.com/path")
	if err != nil {
		t.Fatalf("buildRequest errored: %v", err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", req.Header.Get("Content-Type"))
	}
	if req.URL.RawQuery == "" {
		t.Fatalf("expected query params set")
	}
	if len(data) == 0 {
		t.Fatalf("expected request body bytes returned")
	}
}

func TestResolveBaseURL_PanicOnBadType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on bad target type")
		}
	}()
	resolveBaseURL(123)
}
