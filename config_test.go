package gswag_test

import (
	"testing"

	"github.com/oaswrap/gswag"
)

func TestBearerJWT(t *testing.T) {
	s := gswag.BearerJWT()
	if s.Type != "http" {
		t.Errorf("expected type 'http', got %q", s.Type)
	}
	if s.Scheme != "bearer" {
		t.Errorf("expected scheme 'bearer', got %q", s.Scheme)
	}
	if s.BearerFormat != "JWT" {
		t.Errorf("expected bearerFormat 'JWT', got %q", s.BearerFormat)
	}
}

func TestAPIKeyHeader(t *testing.T) {
	s := gswag.APIKeyHeader("X-API-Key")
	if s.Type != "apiKey" {
		t.Errorf("expected type 'apiKey', got %q", s.Type)
	}
	if s.In != "header" {
		t.Errorf("expected in 'header', got %q", s.In)
	}
	if s.Name != "X-API-Key" {
		t.Errorf("expected name 'X-API-Key', got %q", s.Name)
	}
}

func TestAPIKeyQuery(t *testing.T) {
	s := gswag.APIKeyQuery("api_key")
	if s.Type != "apiKey" {
		t.Errorf("expected type 'apiKey', got %q", s.Type)
	}
	if s.In != "query" {
		t.Errorf("expected in 'query', got %q", s.In)
	}
	if s.Name != "api_key" {
		t.Errorf("expected name 'api_key', got %q", s.Name)
	}
}

func TestAPIKeyCookie(t *testing.T) {
	s := gswag.APIKeyCookie("session")
	if s.Type != "apiKey" {
		t.Errorf("expected type 'apiKey', got %q", s.Type)
	}
	if s.In != "cookie" {
		t.Errorf("expected in 'cookie', got %q", s.In)
	}
	if s.Name != "session" {
		t.Errorf("expected name 'session', got %q", s.Name)
	}
}

func TestOAuth2Implicit(t *testing.T) {
	s := gswag.OAuth2Implicit("https://petstore3.swagger.io/oauth/authorize", map[string]string{
		"write:pets": "modify pets in your account",
		"read:pets":  "read your pets",
	})
	if s.Type != "oauth2" {
		t.Errorf("expected type 'oauth2', got %q", s.Type)
	}
	if s.AuthorizationURL != "https://petstore3.swagger.io/oauth/authorize" {
		t.Errorf("expected auth URL set, got %q", s.AuthorizationURL)
	}
	if len(s.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(s.Scopes))
	}
}

func TestInit_DefaultOutputPath(t *testing.T) {
	gswag.Init(&gswag.Config{
		Title:   "Test",
		Version: "1.0.0",
	})
	// The WriteSpec call should not panic and produce a file at the default path.
	// We validate indirectly by running ValidateSpec without crash.
	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error: %s", iss)
		}
	}
}

func TestInit_DefaultVersion(t *testing.T) {
	gswag.Init(&gswag.Config{
		Title: "Test",
		// Version intentionally empty — should default to "0.1.0"
	})
	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Message == "version is required" {
			t.Errorf("version should have been defaulted: %s", iss)
		}
	}
}

func TestInit_WithSecuritySchemes(t *testing.T) {
	gswag.Init(&gswag.Config{
		Title:   "Secure API",
		Version: "1.0.0",
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"bearerAuth": gswag.BearerJWT(),
			"apiKey":     gswag.APIKeyHeader("X-API-Key"),
		},
	})
	// ValidateSpec should not report undeclared scheme errors for these.
	issues := gswag.ValidateSpec()
	for _, iss := range issues {
		if iss.Severity == "error" {
			t.Errorf("unexpected error after init with schemes: %s", iss)
		}
	}
}
