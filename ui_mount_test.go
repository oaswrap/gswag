package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag"
)

func TestNewSwaggerUIHandler_Mount(t *testing.T) {
	cfg := &gswag.Config{
		Title:   "MountTest",
		Version: "0.0.1",
	}
	gswag.Init(cfg)

	uiCfg := &gswag.UIConfig{
		DocsPath: "/docs",
		SpecPath: "/docs/openapi.json",
	}

	h, err := gswag.NewSwaggerUIHandler(uiCfg)
	if err != nil {
		t.Fatalf("NewSwaggerUIHandler failed: %v", err)
	}

	ts := httptest.NewServer(h)
	defer ts.Close()

	// Docs page
	resp, err := http.Get(ts.URL + uiCfg.DocsPath)
	if err != nil {
		t.Fatalf("GET docs failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for docs, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Spec path
	resp2, err := http.Get(ts.URL + uiCfg.SpecPath)
	if err != nil {
		t.Fatalf("GET spec failed: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for spec, got %d", resp2.StatusCode)
	}
	resp2.Body.Close()
}
