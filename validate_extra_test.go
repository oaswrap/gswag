package gswag

import (
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func TestValidateResponseAgainstOperation_NilBuilder(t *testing.T) {
	_, err := validateResponseAgainstOperation(nil, &recordedResponse{})
	if err == nil {
		t.Fatalf("expected error for nil builder")
	}
}

func TestValidateResponseAgainstOperation_TypedModelUnmarshalFail(t *testing.T) {
	b := newRequestBuilder("GET", "/x")
	b.respBodies[200] = struct{ ID int }{}
	res := &recordedResponse{StatusCode: 200, BodyBytes: []byte("notjson")}

	issues, err := validateResponseAgainstOperation(b, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) == 0 {
		t.Fatalf("expected unmarshal errors returned as issues")
	}
}

func TestValidateResponseAgainstOperation_JSONSchemaValidation(t *testing.T) {
	// Create a collector with a spec that expects {"id": string} for GET /p
	sc := newSpecCollector(&Config{Title: "T", Version: "v"})
	globalCollector = sc
	// Ensure paths map initialized
	if sc.reflector.Spec.Paths.MapOfPathItemValues == nil {
		sc.reflector.Spec.Paths.MapOfPathItemValues = map[string]openapi3.PathItem{}
	}

	// Build an operation entry in the spec manually.
	schema := openapi3.Schema{}
	schema.WithType(openapi3.SchemaTypeObject)
	// property id: string
	prop := openapi3.Schema{}
	prop.WithType(openapi3.SchemaTypeString)
	sref := openapi3.SchemaOrRef{}
	sref.WithSchema(prop)
	schema.WithProperties(map[string]openapi3.SchemaOrRef{"id": sref})
	schema.Required = []string{"id"}

	mt := openapi3.MediaType{Schema: &openapi3.SchemaOrRef{}}
	mt.Schema.WithSchema(schema)

	resp := openapi3.Response{}
	resp.Content = map[string]openapi3.MediaType{"application/json": mt}

	ror := openapi3.ResponseOrRef{}
	ror.WithResponse(resp)

	op := openapi3.Operation{}
	op.Responses = openapi3.Responses{}
	op.Responses.MapOfResponseOrRefValues = map[string]openapi3.ResponseOrRef{"200": ror}

	pi := openapi3.PathItem{}
	pi.MapOfOperationValues = map[string]openapi3.Operation{"get": op}
	sc.reflector.Spec.Paths.MapOfPathItemValues["/p"] = pi

	b := newRequestBuilder("GET", "/p")
	res := &recordedResponse{StatusCode: 200, BodyBytes: []byte(`{"name":"x"}`)}

	issues, err := validateResponseAgainstOperation(b, res)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(issues) == 0 {
		t.Fatalf("expected validation issues when id missing")
	}
}
