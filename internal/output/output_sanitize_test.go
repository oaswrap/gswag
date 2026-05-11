package output_test

import (
	"encoding/json"
	"testing"

	outputpkg "github.com/oaswrap/gswag/internal/output"
	"github.com/oaswrap/spec/openapi"
)

func TestSanitizeSpecForSerialization_FillsMissingParameterIn(t *testing.T) {
	spec := &openapi.Document{Paths: map[string]*openapi.PathItem{
		"/pets/{id}": {
			Get: &openapi.Operation{
				Parameters: []*openapi.Parameter{
					{Name: "id", Required: true},
					{Name: "status"},
				},
			},
		},
	}}

	outputpkg.SanitizeSpecForSerialization(spec)

	op := spec.Paths["/pets/{id}"].Get
	if len(op.Parameters) != 2 {
		t.Fatalf("expected 2 params, got %d", len(op.Parameters))
	}
	id := op.Parameters[0]
	if id.In != "path" {
		t.Fatalf("expected id param location path, got %q", id.In)
	}
	if !id.Required {
		t.Fatalf("expected id path param to be required")
	}
	status := op.Parameters[1]
	if status.In != "query" {
		t.Fatalf("expected status param location query, got %q", status.In)
	}
	if _, err := json.Marshal(spec); err != nil {
		t.Fatalf("expected sanitized spec to marshal, got error: %v", err)
	}
}

func TestSanitizeSpecForSerialization_DedupesParametersByNameAndLocation(t *testing.T) {
	spec := &openapi.Document{Paths: map[string]*openapi.PathItem{
		"/pets": {
			Get: &openapi.Operation{
				Parameters: []*openapi.Parameter{
					{Name: "status", In: "query"},
					{Name: "status", In: "query"},
				},
			},
		},
	}}

	outputpkg.SanitizeSpecForSerialization(spec)

	op := spec.Paths["/pets"].Get
	if len(op.Parameters) != 1 {
		t.Fatalf("expected deduped params length 1, got %d", len(op.Parameters))
	}
}
