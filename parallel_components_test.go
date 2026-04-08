package gswag

import (
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func TestMergeSpec_SecuritySchemes(t *testing.T) {
	base := &openapi3.Spec{}
	base.Components = &openapi3.Components{}
	base.Components.SecuritySchemesEns()
	base.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem("Existing", openapi3.SecuritySchemeOrRef{})

	src := &openapi3.Spec{}
	src.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}
	src.Components = &openapi3.Components{}
	src.Components.SecuritySchemesEns()
	src.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem("New", openapi3.SecuritySchemeOrRef{})
	src.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem("Existing", openapi3.SecuritySchemeOrRef{})

	mergeSpec(base, src)

	if base.Components == nil || base.Components.SecuritySchemes == nil {
		t.Fatalf("security schemes missing after merge")
	}
	if _, ok := base.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues["Existing"]; !ok {
		t.Fatalf("Existing scheme lost")
	}
	if _, ok := base.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues["New"]; !ok {
		t.Fatalf("New scheme not added")
	}
}

func TestMergeSpec_SrcComponentsNil(t *testing.T) {
	base := &openapi3.Spec{}
	base.Components = &openapi3.Components{}
	base.Components.SchemasEns()
	base.Components.Schemas.WithMapOfSchemaOrRefValuesItem("A", openapi3.SchemaOrRef{})

	src := &openapi3.Spec{} // src.Components == nil
	src.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}

	mergeSpec(base, src)

	if base.Components == nil || base.Components.Schemas == nil {
		t.Fatalf("components/schemas lost after merging nil src")
	}
	if _, ok := base.Components.Schemas.MapOfSchemaOrRefValues["A"]; !ok {
		t.Fatalf("Existing schema A lost")
	}
}

func TestMergeSpec_DstComponentsNil(t *testing.T) {
	base := &openapi3.Spec{} // components nil

	src := &openapi3.Spec{}
	src.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}
	src.Components = &openapi3.Components{}
	src.Components.SecuritySchemesEns()
	src.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem("S", openapi3.SecuritySchemeOrRef{})

	mergeSpec(base, src)

	if base.Components == nil || base.Components.SecuritySchemes == nil {
		t.Fatalf("components or security schemes missing after merge into nil dst")
	}
	if _, ok := base.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues["S"]; !ok {
		t.Fatalf("S scheme not added to dst")
	}
}

// TestMergeSpec_SrcNoPaths verifies that component schemas are merged even when
// src has no paths (previously the function returned early and skipped schemas).
func TestMergeSpec_SrcNoPaths(t *testing.T) {
	base := &openapi3.Spec{}
	base.Components = &openapi3.Components{}
	base.Components.SchemasEns()
	base.Components.Schemas.WithMapOfSchemaOrRefValuesItem("Base", openapi3.SchemaOrRef{})

	// src has no paths at all (MapOfPathItemValues is nil) but contributes a schema.
	src := &openapi3.Spec{}
	src.Components = &openapi3.Components{}
	src.Components.SchemasEns()
	src.Components.Schemas.WithMapOfSchemaOrRefValuesItem("FromSrc", openapi3.SchemaOrRef{})

	mergeSpec(base, src)

	if base.Components == nil || base.Components.Schemas == nil {
		t.Fatalf("schemas missing after merge")
	}
	if _, ok := base.Components.Schemas.MapOfSchemaOrRefValues["Base"]; !ok {
		t.Fatalf("Base schema lost")
	}
	if _, ok := base.Components.Schemas.MapOfSchemaOrRefValues["FromSrc"]; !ok {
		t.Fatalf("FromSrc schema not merged despite src having no paths")
	}
}

// TestMergeSpec_ExtendedComponents verifies that Responses, Parameters,
// RequestBodies, Headers, Examples, Links, and Callbacks are merged.
func TestMergeSpec_ExtendedComponents(t *testing.T) {
	base := &openapi3.Spec{}

	src := &openapi3.Spec{}
	src.Components = &openapi3.Components{}
	src.Components.ResponsesEns().WithMapOfResponseOrRefValuesItem("NotFound", openapi3.ResponseOrRef{})
	src.Components.ParametersEns().WithMapOfParameterOrRefValuesItem("LimitParam", openapi3.ParameterOrRef{})
	src.Components.RequestBodiesEns().WithMapOfRequestBodyOrRefValuesItem("CreateBody", openapi3.RequestBodyOrRef{})
	src.Components.HeadersEns().WithMapOfHeaderOrRefValuesItem("X-Rate-Limit", openapi3.HeaderOrRef{})
	src.Components.ExamplesEns().WithMapOfExampleOrRefValuesItem("FooExample", openapi3.ExampleOrRef{})
	src.Components.LinksEns().WithMapOfLinkOrRefValuesItem("UserLink", openapi3.LinkOrRef{})
	src.Components.CallbacksEns().WithMapOfCallbackOrRefValuesItem("OnEvent", openapi3.CallbackOrRef{})

	mergeSpec(base, src)

	if base.Components == nil {
		t.Fatal("components missing after merge")
	}
	checks := []struct {
		name string
		ok   bool
	}{
		{"Responses.NotFound", base.Components.Responses != nil && len(base.Components.Responses.MapOfResponseOrRefValues) > 0},
		{"Parameters.LimitParam", base.Components.Parameters != nil && len(base.Components.Parameters.MapOfParameterOrRefValues) > 0},
		{"RequestBodies.CreateBody", base.Components.RequestBodies != nil && len(base.Components.RequestBodies.MapOfRequestBodyOrRefValues) > 0},
		{"Headers.X-Rate-Limit", base.Components.Headers != nil && len(base.Components.Headers.MapOfHeaderOrRefValues) > 0},
		{"Examples.FooExample", base.Components.Examples != nil && len(base.Components.Examples.MapOfExampleOrRefValues) > 0},
		{"Links.UserLink", base.Components.Links != nil && len(base.Components.Links.MapOfLinkOrRefValues) > 0},
		{"Callbacks.OnEvent", base.Components.Callbacks != nil && len(base.Components.Callbacks.MapOfCallbackOrRefValues) > 0},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("component %s not merged", c.name)
		}
	}
}
