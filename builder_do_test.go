package gswag

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBuilder_Do_SuccessRegisters(t *testing.T) {
	// setup server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1}`))
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
		w.Write([]byte("notjson"))
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
				if strings.Contains(r.(string), "response does not match declared schema") || strings.Contains(r.(string), "response validation error") {
					didPanic = true
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
				if strings.Contains(r.(string), "HTTP request failed") {
					didPanic = true
				}
			}
		}()
		_ = b.do(urlSrv)
	}()
	if !didPanic {
		t.Fatalf("expected network failure to panic")
	}
}
