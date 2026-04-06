package gswag

import (
	"net/http/httptest"
	"testing"
)

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
