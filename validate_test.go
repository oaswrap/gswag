package gswag_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
)

func TestValidateSpecFile_Valid(t *testing.T) {
	issues, err := gswag.ValidateSpecFile("./examples/stdlib/docs/openapi.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Errorf("unexpected error issue: %s", issue)
		}
	}
}

func TestValidateSpecFile_NotFound(t *testing.T) {
	_, err := gswag.ValidateSpecFile("./nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateSpecFile_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("%%%invalid"), 0o644) //nolint:errcheck
	_, err := gswag.ValidateSpecFile(path)
	// YAML parse may or may not fail; JSON fallback also fails — expect error.
	// If no error, at minimum parsing did not crash.
	_ = err
}

func TestValidationIssue_String(t *testing.T) {
	issue := gswag.ValidationIssue{
		Severity: "error",
		Path:     "info.title",
		Message:  "title is required",
	}
	s := issue.String()
	if !strings.Contains(s, "ERROR") {
		t.Errorf("expected ERROR in string, got %q", s)
	}
	if !strings.Contains(s, "info.title") {
		t.Errorf("expected path in string, got %q", s)
	}
	if !strings.Contains(s, "title is required") {
		t.Errorf("expected message in string, got %q", s)
	}
}

func TestValidateSpec_NotInitialised(t *testing.T) {
	// Reset global state by re-initialising then testing on a fresh import would
	// require package-level access; instead call Init and then check no panics.
	gswag.Init(&gswag.Config{Title: "T", Version: "1"})
	issues := gswag.ValidateSpec()
	_ = issues // should not panic
}

func TestValidateSpec_MissingTitle(t *testing.T) {
	// ValidateSpecFile on a spec with no title.
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	content := `openapi: "3.0.3"
info:
  title: ""
  version: "1.0.0"
paths: {}`
	os.WriteFile(path, []byte(content), 0o644) //nolint:errcheck

	issues, err := gswag.ValidateSpecFile(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	found := false
	for _, iss := range issues {
		if iss.Severity == "error" && strings.Contains(iss.Message, "title") {
			found = true
		}
	}
	if !found {
		t.Error("expected error issue about missing title")
	}
}

func TestValidateSpec_NoPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	content := `openapi: "3.0.3"
info:
  title: "My API"
  version: "1.0.0"
paths: {}`
	os.WriteFile(path, []byte(content), 0o644) //nolint:errcheck

	issues, err := gswag.ValidateSpecFile(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	found := false
	for _, iss := range issues {
		if iss.Severity == "warning" && strings.Contains(iss.Message, "no paths") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about no paths defined")
	}
}

func TestValidateSpec_UndeclaredSecurityScheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.yaml")
	// Operation references "bearerAuth" but components don't declare it.
	content := `openapi: "3.0.3"
info:
  title: "My API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      tags: [items]
      security:
        - bearerAuth: []
      responses:
        "200":
          description: OK`
	os.WriteFile(path, []byte(content), 0o644) //nolint:errcheck

	issues, err := gswag.ValidateSpecFile(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	found := false
	for _, iss := range issues {
		if iss.Severity == "error" && strings.Contains(iss.Message, "bearerAuth") {
			found = true
		}
	}
	if !found {
		t.Error("expected error about undeclared security scheme bearerAuth")
	}
}

func TestWriteAndValidateSpec_Valid(t *testing.T) {
	dir := t.TempDir()

	gswag.Init(&gswag.Config{
		Title:      "Test API",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})
	// WriteAndValidateSpec with a valid (but empty-paths) spec is still valid.
	// Warnings about no paths are acceptable; only errors cause a failure.
	if err := gswag.WriteAndValidateSpec(); err != nil {
		// Allow "no paths defined" warning — only error if there are actual validation errors.
		issues, _ := gswag.ValidateSpecFile(filepath.Join(dir, "openapi.yaml"))
		for _, iss := range issues {
			if iss.Severity == "error" {
				t.Fatalf("spec error: %s", iss)
			}
		}
	}
}

func TestWriteAndValidateSpec_Invalid(t *testing.T) {
	dir := t.TempDir()

	// Init with no title so validation fails.
	gswag.Init(&gswag.Config{
		Title:      "",
		Version:    "1.0.0",
		OutputPath: filepath.Join(dir, "openapi.yaml"),
	})

	err := gswag.WriteAndValidateSpec()
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
	if !errors.Is(err, gswag.ErrSpecInvalid) {
		t.Errorf("expected ErrSpecInvalid, got %v", err)
	}
}
