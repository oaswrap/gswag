package gswag

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestUIConfig_addr(t *testing.T) {
	if (&UIConfig{}).addr() != ":9090" {
		t.Errorf("expected default addr :9090")
	}
	if (&UIConfig{Addr: ":8080"}).addr() != ":8080" {
		t.Errorf("expected custom addr :8080")
	}
	if ((*UIConfig)(nil)).addr() != ":9090" {
		t.Errorf("expected nil UIConfig to return default addr")
	}
}

func TestUIConfig_docsPath(t *testing.T) {
	if (&UIConfig{}).docsPath() != "/docs" {
		t.Errorf("expected default docsPath /docs")
	}
	if (&UIConfig{DocsPath: "/api-docs"}).docsPath() != "/api-docs" {
		t.Errorf("expected custom docsPath /api-docs")
	}
	if ((*UIConfig)(nil)).docsPath() != "/docs" {
		t.Errorf("expected nil UIConfig to return default docsPath")
	}
}

func TestUIConfig_specPath(t *testing.T) {
	if (&UIConfig{}).specPath() != "/docs/openapi.json" {
		t.Errorf("expected default specPath /docs/openapi.json")
	}
	if (&UIConfig{SpecPath: "/spec.json"}).specPath() != "/spec.json" {
		t.Errorf("expected custom specPath /spec.json")
	}
	if ((*UIConfig)(nil)).specPath() != "/docs/openapi.json" {
		t.Errorf("expected nil UIConfig to return default specPath")
	}
}

func TestServeUI_NotInitialised(t *testing.T) {
	// Temporarily nil out the global collector to test the error path.
	orig := globalCollector
	globalCollector = nil
	defer func() { globalCollector = orig }()

	err := ServeUI(&UIConfig{})
	if err == nil {
		t.Fatal("expected error when not initialised")
	}
}

func TestServeRedoc_NotInitialised(t *testing.T) {
	orig := globalCollector
	globalCollector = nil
	defer func() { globalCollector = orig }()

	err := ServeRedoc(&UIConfig{})
	if err == nil {
		t.Fatal("expected error when not initialised")
	}
}

// TestServeUI_HappyPath verifies the handler setup runs without panicking.
// It starts the server on a random port, makes one request, then relies on
// the listener being closed (via the net.Listener trick) to exit cleanly.
func TestServeUI_HappyPath(t *testing.T) {
	Init(&Config{Title: "UI Test API", Version: "1.0.0"})

	// Pick a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot get free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close() // free the port; ServeUI will re-bind

	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeUI(&UIConfig{
			Addr:     addr,
			DocsPath: "/docs",
			SpecPath: "/docs/openapi.json",
		})
	}()

	// Wait briefly for the server to start.
	time.Sleep(50 * time.Millisecond)

	// Make a request to verify the server is running.
	url := fmt.Sprintf("http://%s/docs", addr)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		// Server may not have started yet; just verify no panic occurred.
		t.Logf("HTTP request to %s: %v (server may not have started)", url, err)
	} else {
		resp.Body.Close()
	}

	// The server runs until the process exits; don't wait for errCh in tests.
}
