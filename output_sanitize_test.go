package gswag

import (
	"encoding/json"
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func TestSanitizeSpecForSerialization_FillsMissingParameterIn(t *testing.T) {
	req := true
	spec := &openapi3.Spec{}
	spec.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{
		"/pets/{id}": {
			MapOfOperationValues: map[string]openapi3.Operation{
				"get": {
					Parameters: []openapi3.ParameterOrRef{
						{Parameter: &openapi3.Parameter{Name: "id", In: openapi3.ParameterIn(""), Required: &req}},
						{Parameter: &openapi3.Parameter{Name: "status", In: openapi3.ParameterIn("")}},
					},
				},
			},
		},
	}}

	sanitizeSpecForSerialization(spec)

	op := spec.Paths.MapOfPathItemValues["/pets/{id}"].MapOfOperationValues["get"]
	if len(op.Parameters) != 2 {
		t.Fatalf("expected 2 params, got %d", len(op.Parameters))
	}

	id := op.Parameters[0].Parameter
	if string(id.In) != "path" {
		t.Fatalf("expected id param location path, got %q", id.In)
	}
	if id.Required == nil || !*id.Required {
		t.Fatalf("expected id path param to be required")
	}

	status := op.Parameters[1].Parameter
	if string(status.In) != "query" {
		t.Fatalf("expected status param location query, got %q", status.In)
	}

	if _, err := json.Marshal(spec); err != nil {
		t.Fatalf("expected sanitized spec to marshal, got error: %v", err)
	}
}

func TestSanitizeSpecForSerialization_DedupesParametersByNameAndLocation(t *testing.T) {
	spec := &openapi3.Spec{}
	spec.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{
		"/pets": {
			MapOfOperationValues: map[string]openapi3.Operation{
				"get": {
					Parameters: []openapi3.ParameterOrRef{
						{Parameter: &openapi3.Parameter{Name: "status", In: openapi3.ParameterIn("query")}},
						{Parameter: &openapi3.Parameter{Name: "status", In: openapi3.ParameterIn("query")}},
					},
				},
			},
		},
	}}

	sanitizeSpecForSerialization(spec)

	op := spec.Paths.MapOfPathItemValues["/pets"].MapOfOperationValues["get"]
	if len(op.Parameters) != 1 {
		t.Fatalf("expected deduped params length 1, got %d", len(op.Parameters))
	}
}
