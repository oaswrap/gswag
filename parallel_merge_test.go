package gswag

import (
	"testing"

	"github.com/oaswrap/spec/openapi"
)

func TestMergeSpecPathsAndSchemas(t *testing.T) {
	base := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/p": {Get: &openapi.Operation{Responses: map[string]*openapi.Response{}}},
		},
		Components: &openapi.Components{Schemas: map[string]*openapi.Schema{"Existing": {}}},
	}
	src := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/p": {
				Get:  &openapi.Operation{Responses: map[string]*openapi.Response{}},
				Post: &openapi.Operation{Responses: map[string]*openapi.Response{}},
			},
		},
		Components: &openapi.Components{Schemas: map[string]*openapi.Schema{"New": {}, "Existing": {}}},
	}

	mergeSpec(base, src)

	pi := base.Paths["/p"]
	if pi == nil || pi.Get == nil {
		t.Fatalf("expected get operation preserved")
	}
	if pi.Post == nil {
		t.Fatalf("expected post operation merged")
	}
	if _, ok := base.Components.Schemas["Existing"]; !ok {
		t.Fatalf("expected existing schema preserved")
	}
	if _, ok := base.Components.Schemas["New"]; !ok {
		t.Fatalf("expected new schema merged")
	}
}

func TestMergeSpecSecuritySchemes(t *testing.T) {
	base := &openapi.Document{
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{"Existing": {Type: "http"}},
		},
	}
	src := &openapi.Document{
		Paths: map[string]*openapi.PathItem{},
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{"New": {Type: "apiKey"}, "Existing": {Type: "apiKey"}},
		},
	}

	mergeSpec(base, src)

	if _, ok := base.Components.SecuritySchemes["Existing"]; !ok {
		t.Fatalf("expected existing scheme preserved")
	}
	if _, ok := base.Components.SecuritySchemes["New"]; !ok {
		t.Fatalf("expected new scheme merged")
	}
}

func TestMergeSpecNoSrcComponents(t *testing.T) {
	base := &openapi.Document{
		Paths:      map[string]*openapi.PathItem{},
		Components: &openapi.Components{Schemas: map[string]*openapi.Schema{"A": {}}},
	}
	src := &openapi.Document{Paths: map[string]*openapi.PathItem{}}

	mergeSpec(base, src)

	if _, ok := base.Components.Schemas["A"]; !ok {
		t.Fatalf("expected base schema to remain")
	}
}

func TestMergeSpecNilDstComponents(t *testing.T) {
	base := &openapi.Document{}
	src := &openapi.Document{
		Paths:      map[string]*openapi.PathItem{},
		Components: &openapi.Components{SecuritySchemes: map[string]*openapi.SecurityScheme{"S": {Type: "http"}}},
	}

	mergeSpec(base, src)

	if base.Components == nil || base.Components.SecuritySchemes["S"] == nil {
		t.Fatalf("expected components to be initialized and merged")
	}
}

func TestMergeSpecMergesComponentsEvenWhenSrcHasNoPaths(t *testing.T) {
	base := &openapi.Document{Components: &openapi.Components{Schemas: map[string]*openapi.Schema{"Base": {}}}}
	src := &openapi.Document{Components: &openapi.Components{Schemas: map[string]*openapi.Schema{"FromSrc": {}}}}

	mergeSpec(base, src)

	if _, ok := base.Components.Schemas["Base"]; !ok {
		t.Fatalf("expected base schema to remain")
	}
	if _, ok := base.Components.Schemas["FromSrc"]; !ok {
		t.Fatalf("expected source schema to merge even without paths")
	}
}

func TestMergeSpecAllComponentKinds(t *testing.T) {
	base := &openapi.Document{}
	src := &openapi.Document{Components: &openapi.Components{
		Responses:     map[string]*openapi.Response{"NotFound": {}},
		Parameters:    map[string]*openapi.Parameter{"LimitParam": {}},
		RequestBodies: map[string]*openapi.RequestBody{"CreateBody": {}},
		Headers:       map[string]*openapi.Header{"X-Rate-Limit": {}},
		Examples:      map[string]*openapi.Example{"FooExample": {}},
		Links:         map[string]*openapi.Link{"UserLink": {}},
		Callbacks:     map[string]*openapi.Callback{"OnEvent": {}},
	}}

	mergeSpec(base, src)

	checks := []struct {
		name string
		ok   bool
	}{
		{"responses", len(base.Components.Responses) > 0},
		{"parameters", len(base.Components.Parameters) > 0},
		{"requestBodies", len(base.Components.RequestBodies) > 0},
		{"headers", len(base.Components.Headers) > 0},
		{"examples", len(base.Components.Examples) > 0},
		{"links", len(base.Components.Links) > 0},
		{"callbacks", len(base.Components.Callbacks) > 0},
	}
	for _, check := range checks {
		if !check.ok {
			t.Fatalf("expected %s to be merged", check.name)
		}
	}
}
