package gswag

import (
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func TestMergeSpec_PathsAndSchemas(t *testing.T) {
	// base spec has /p GET and schema Existing
	base := &openapi3.Spec{}
	base.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}
	base.Paths.MapOfPathItemValues["/p"] = openapi3.PathItem{
		MapOfOperationValues: map[string]openapi3.Operation{"get": {}},
	}

	base.Components = &openapi3.Components{}
	base.Components.SchemasEns()
	base.Components.Schemas.WithMapOfSchemaOrRefValuesItem("Existing", openapi3.SchemaOrRef{})

	// src spec has /p GET (should not overwrite) and POST (should be added)
	src := &openapi3.Spec{}
	src.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}
	src.Paths.MapOfPathItemValues["/p"] = openapi3.PathItem{
		MapOfOperationValues: map[string]openapi3.Operation{"get": {}, "post": {}},
	}

	src.Components = &openapi3.Components{}
	src.Components.SchemasEns()
	src.Components.Schemas.WithMapOfSchemaOrRefValuesItem("New", openapi3.SchemaOrRef{})
	src.Components.Schemas.WithMapOfSchemaOrRefValuesItem("Existing", openapi3.SchemaOrRef{})

	mergeSpec(base, src)

	// paths
	pi, ok := base.Paths.MapOfPathItemValues["/p"]
	if !ok {
		t.Fatalf("path /p missing")
	}
	if _, hasGet := pi.MapOfOperationValues["get"]; !hasGet {
		t.Fatalf("get missing after merge")
	}
	if _, hasPost := pi.MapOfOperationValues["post"]; !hasPost {
		t.Fatalf("post not merged")
	}

	// schemas: Existing should remain, New should be added
	if base.Components == nil || base.Components.Schemas == nil {
		t.Fatalf("components schemas missing")
	}
	if _, ok := base.Components.Schemas.MapOfSchemaOrRefValues["Existing"]; !ok {
		t.Fatalf("Existing schema lost")
	}
	if _, ok := base.Components.Schemas.MapOfSchemaOrRefValues["New"]; !ok {
		t.Fatalf("New schema not added")
	}
}
