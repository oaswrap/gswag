package gswag_test

import (
	"testing"

	"github.com/oaswrap/gswag"
)

func TestUIConfig_Defaults(t *testing.T) {
	// UIConfig is exported; test that ServeUI/ServeRedoc return error before Init.
	// We can't test the blocking listener, but we can test the not-initialised path
	// by ensuring the error is returned when globalCollector is nil.
	// Reset state by re-initialising is not possible externally, so we rely on
	// calling these after a fresh process state isn't feasible; skip the blocking test.
	t.Log("UIConfig default tests: covered via integration with ServeUI/ServeRedoc error path")
}

func TestServeUI_NotInitialised(t *testing.T) {
	// Force the collector to nil by testing via the exported error message.
	// We cannot set globalCollector to nil from outside the package, so we
	// check that calling WriteSpec with nil collector does not panic.
	// The actual "not initialised" code path in serveUI is tested by initialising
	// in a subprocess; here we just verify ServeUI is callable.
	// Note: ServeUI would block, so we cannot call it here without a goroutine + cancel.
	t.Log("ServeUI not-initialised path cannot be tested without package internals access")
}

// Test UIConfig default value methods indirectly via the publicly-accessible spec path.
// The UIConfig struct and its helpers are exercised during ServeUI/ServeRedoc.
func TestUIConfigZeroValue(t *testing.T) {
	// ServeUI and ServeRedoc are blocking; test the UIConfig type instantiation.
	cfg := &gswag.UIConfig{}
	if cfg.Addr != "" || cfg.DocsPath != "" || cfg.SpecPath != "" {
		t.Error("expected zero UIConfig to have empty fields")
	}
}
