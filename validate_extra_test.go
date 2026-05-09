package gswag

import (
	"testing"

	"github.com/oaswrap/spec/openapi"
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
	sc := newSpecCollector(&Config{Title: "T", Version: "v"})
	globalCollector = sc

	schema := &openapi.Schema{
		Type:     "object",
		Required: []string{"id"},
		Properties: map[string]*openapi.Schema{
			"id": {Type: "string"},
		},
	}
	op := &openapi.Operation{Responses: map[string]*openapi.Response{
		"200": {Content: map[string]openapi.MediaType{applicationJSON: {Schema: schema}}},
	}}
	sc.doc.Paths["/p"] = &openapi.PathItem{Get: op}

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
