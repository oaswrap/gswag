package schemautil_test

import (
	"testing"

	"github.com/oaswrap/gswag/internal/schemautil"
)

func TestInferSchema_Object(t *testing.T) {
	data := []byte(`{"id":1,"name":"Alice","active":true,"score":9.5}`)
	sor := schemautil.InferSchema(data)
	if sor == nil || sor.Schema == nil {
		t.Fatal("expected non-nil schema")
	}
	s := sor.Schema
	if s.Type == nil || string(*s.Type) != "object" {
		t.Fatalf("expected type object, got %v", s.Type)
	}
	if len(s.Properties) != 4 {
		t.Fatalf("expected 4 properties, got %d", len(s.Properties))
	}
}

func TestInferSchema_Array(t *testing.T) {
	data := []byte(`[{"id":1},{"id":2}]`)
	sor := schemautil.InferSchema(data)
	if sor == nil || sor.Schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if string(*sor.Schema.Type) != "array" {
		t.Fatalf("expected type array, got %v", sor.Schema.Type)
	}
}

func TestInferSchema_Empty(t *testing.T) {
	if schemautil.InferSchema(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
	if schemautil.InferSchema([]byte{}) != nil {
		t.Fatal("expected nil for empty bytes")
	}
}

func TestInferSchema_InvalidJSON(t *testing.T) {
	if schemautil.InferSchema([]byte(`not json`)) != nil {
		t.Fatal("expected nil for invalid JSON")
	}
}

func TestInferSchema_Integer(t *testing.T) {
	data := []byte(`{"count":42}`)
	sor := schemautil.InferSchema(data)
	prop := sor.Schema.Properties["count"]
	if string(*prop.Schema.Type) != "integer" {
		t.Fatalf("expected integer, got %v", prop.Schema.Type)
	}
}

func TestInferSchema_Number(t *testing.T) {
	data := []byte(`{"price":9.99}`)
	sor := schemautil.InferSchema(data)
	prop := sor.Schema.Properties["price"]
	if string(*prop.Schema.Type) != "number" {
		t.Fatalf("expected number, got %v", prop.Schema.Type)
	}
}
