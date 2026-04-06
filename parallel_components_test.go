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
