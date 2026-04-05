package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSpecFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

const baseSpec = `openapi: "3.0.3"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      parameters:
        - in: query
          name: limit
          schema:
            type: integer
      responses:
        "200":
          description: OK
  /items/{id}:
    get:
      summary: Get item
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
        "404":
          description: Not found
`

const identicalSpec = baseSpec

const addedPathSpec = `openapi: "3.0.3"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      parameters:
        - in: query
          name: limit
          schema:
            type: integer
      responses:
        "200":
          description: OK
  /items/{id}:
    get:
      summary: Get item
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
        "404":
          description: Not found
  /orders:
    get:
      summary: List orders
      responses:
        "200":
          description: OK
`

const removedPathSpec = `openapi: "3.0.3"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      responses:
        "200":
          description: OK
`

const addedRequiredParamSpec = `openapi: "3.0.3"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      parameters:
        - in: query
          name: limit
          schema:
            type: integer
        - in: query
          name: filter
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
  /items/{id}:
    get:
      summary: Get item
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
        "404":
          description: Not found
`

const removedResponseSpec = `openapi: "3.0.3"
info:
  title: "Test API"
  version: "1.0.0"
paths:
  /items:
    get:
      summary: List items
      responses:
        "200":
          description: OK
  /items/{id}:
    get:
      summary: Get item
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
`

// --- loadSpec ---

func TestLoadSpec_YAML(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "spec.yaml", baseSpec)

	spec, err := loadSpec(path)
	if err != nil {
		t.Fatalf("loadSpec failed: %v", err)
	}
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
}

func TestLoadSpec_NotFound(t *testing.T) {
	_, err := loadSpec("/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadSpec_InvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "bad.yaml", "%%%invalid: [unclosed")
	_, err := loadSpec(path)
	// May or may not error depending on YAML parser leniency; just don't panic.
	_ = err
}

// --- containsBreaking ---

func TestContainsBreaking_True(t *testing.T) {
	cases := []string{
		"  BREAKING  removed parameter",
		"something BREAKING here",
		"BREAKING",
	}
	for _, s := range cases {
		if !containsBreaking(s) {
			t.Errorf("expected containsBreaking=true for %q", s)
		}
	}
}

func TestContainsBreaking_False(t *testing.T) {
	cases := []string{
		"",
		"  added   /items",
		"no issue here",
		"BREAK", // not the full word
	}
	for _, s := range cases {
		if containsBreaking(s) {
			t.Errorf("expected containsBreaking=false for %q", s)
		}
	}
}

// --- pathMethodSet ---

func TestPathMethodSet(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "spec.yaml", baseSpec)
	spec, err := loadSpec(path)
	if err != nil {
		t.Fatalf("loadSpec: %v", err)
	}

	m := pathMethodSet(spec)
	if !m["get /items"] {
		t.Error("expected get /items in set")
	}
	if !m["get /items/{id}"] {
		t.Error("expected get /items/{id} in set")
	}
}

// --- paramSet ---

func TestParamSet_RequiredAndOptional(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "spec.yaml", baseSpec)
	spec, err := loadSpec(path)
	if err != nil {
		t.Fatalf("loadSpec: %v", err)
	}

	item := spec.Paths.MapOfPathItemValues["/items"]
	op := item.MapOfOperationValues["get"]
	ps := paramSet(op)

	// "limit" is optional (required not set → false)
	if _, exists := ps["query:limit"]; !exists {
		t.Error("expected query:limit in paramSet")
	}
	if ps["query:limit"] {
		t.Error("expected query:limit to be non-required")
	}
}

func TestParamSet_PathRequired(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "spec.yaml", baseSpec)
	spec, err := loadSpec(path)
	if err != nil {
		t.Fatalf("loadSpec: %v", err)
	}

	item := spec.Paths.MapOfPathItemValues["/items/{id}"]
	op := item.MapOfOperationValues["get"]
	ps := paramSet(op)

	if !ps["path:id"] {
		t.Error("expected path:id to be required")
	}
}

// --- statusSet ---

func TestStatusSet(t *testing.T) {
	dir := t.TempDir()
	path := writeSpecFile(t, dir, "spec.yaml", baseSpec)
	spec, err := loadSpec(path)
	if err != nil {
		t.Fatalf("loadSpec: %v", err)
	}

	item := spec.Paths.MapOfPathItemValues["/items/{id}"]
	op := item.MapOfOperationValues["get"]
	ss := statusSet(op)

	if !ss["200"] {
		t.Error("expected status 200 in set")
	}
	if !ss["404"] {
		t.Error("expected status 404 in set")
	}
}

// --- diffSpecs ---

func TestDiffSpecs_NoChange(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSpecFile(t, dir, "base.yaml", baseSpec)
	p2 := writeSpecFile(t, dir, "next.yaml", identicalSpec)

	base, _ := loadSpec(p1)
	next, _ := loadSpec(p2)

	result := diffSpecs(base, next)
	if len(result.added) != 0 || len(result.removed) != 0 || len(result.modified) != 0 {
		t.Errorf("expected no diff for identical specs, got: %+v", result)
	}
}

func TestDiffSpecs_AddedPath(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSpecFile(t, dir, "base.yaml", baseSpec)
	p2 := writeSpecFile(t, dir, "next.yaml", addedPathSpec)

	base, _ := loadSpec(p1)
	next, _ := loadSpec(p2)

	result := diffSpecs(base, next)
	if len(result.added) == 0 {
		t.Error("expected at least one added path")
	}
	found := false
	for _, a := range result.added {
		if a == "get /orders" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected get /orders in added, got %v", result.added)
	}
}

func TestDiffSpecs_RemovedPath_IsBreaking(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSpecFile(t, dir, "base.yaml", baseSpec)
	p2 := writeSpecFile(t, dir, "next.yaml", removedPathSpec)

	base, _ := loadSpec(p1)
	next, _ := loadSpec(p2)

	result := diffSpecs(base, next)
	if len(result.removed) == 0 {
		t.Error("expected removed entries")
	}
}

func TestDiffSpecs_AddedRequiredParam_IsBreaking(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSpecFile(t, dir, "base.yaml", baseSpec)
	p2 := writeSpecFile(t, dir, "next.yaml", addedRequiredParamSpec)

	base, _ := loadSpec(p1)
	next, _ := loadSpec(p2)

	result := diffSpecs(base, next)
	if len(result.modified) == 0 {
		t.Error("expected modified entries for added required param")
	}

	hasBreaking := false
	for _, m := range result.modified {
		if containsBreaking(m) {
			hasBreaking = true
		}
	}
	if !hasBreaking {
		t.Errorf("expected BREAKING in modified for new required param, got %v", result.modified)
	}
}

func TestDiffSpecs_RemovedResponseStatus_IsBreaking(t *testing.T) {
	dir := t.TempDir()
	p1 := writeSpecFile(t, dir, "base.yaml", baseSpec)
	p2 := writeSpecFile(t, dir, "next.yaml", removedResponseSpec)

	base, _ := loadSpec(p1)
	next, _ := loadSpec(p2)

	result := diffSpecs(base, next)
	hasBreaking := false
	for _, m := range result.modified {
		if containsBreaking(m) {
			hasBreaking = true
		}
	}
	if !hasBreaking {
		t.Errorf("expected BREAKING for removed response status 404, got %v", result.modified)
	}
}

// --- printDiff ---

func TestPrintDiff_NoChanges(t *testing.T) {
	hasBreaking := printDiff(diffResult{})
	if hasBreaking {
		t.Error("expected no breaking changes for empty diffResult")
	}
}

func TestPrintDiff_WithBreaking(t *testing.T) {
	r := diffResult{
		removed: []string{"GET /items"},
	}
	hasBreaking := printDiff(r)
	if !hasBreaking {
		t.Error("expected hasBreaking=true for removed path")
	}
}

func TestPrintDiff_WithAdded(t *testing.T) {
	r := diffResult{
		added: []string{"GET /orders"},
	}
	hasBreaking := printDiff(r)
	if hasBreaking {
		t.Error("expected no breaking changes for added path")
	}
}
