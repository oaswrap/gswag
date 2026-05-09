package gswag

import (
	"testing"
)

func TestOpenAPIVersionSupport(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{OpenAPI300, "3.0.0"},
		{OpenAPI301, "3.0.1"},
		{OpenAPI302, "3.0.2"},
		{OpenAPI303, "3.0.3"},
		{OpenAPI304, "3.0.4"},
		{OpenAPI310, "3.1.0"},
		{OpenAPI311, "3.1.1"},
		{OpenAPI312, "3.1.2"},
		{OpenAPI320, "3.2.0"},
		{"", "3.0.3"}, // Default
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			cfg := &Config{
				Title:          "Test",
				Version:        "1.0.0",
				OpenAPIVersion: tc.version,
			}
			sc := newSpecCollector(cfg)
			if sc.doc.OpenAPI != tc.expected {
				t.Errorf("expected version %q, got %q", tc.expected, sc.doc.OpenAPI)
			}
		})
	}
}

func TestBuildOpenAPIOptions_UsesConfigVersion(t *testing.T) {
	cfg := &Config{
		OpenAPIVersion: OpenAPI310,
	}
	// Verify indirectly via newSpecCollector.
	sc := newSpecCollector(cfg)
	if sc.doc.OpenAPI != "3.1.0" {
		t.Errorf("expected version 3.1.0, got %q", sc.doc.OpenAPI)
	}
}
