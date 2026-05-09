package schemautil_test

import (
	"testing"

	"github.com/oaswrap/gswag/internal/schemautil"
)

func TestInferSchema_Object(t *testing.T) {
	data := []byte(`{"id":1,"name":"Alice","active":true,"score":9.5}`)
	s := schemautil.InferSchema(data)
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if s.Type != "object" {
		t.Fatalf("expected type object, got %v", s.Type)
	}
	if len(s.Properties) != 4 {
		t.Fatalf("expected 4 properties, got %d", len(s.Properties))
	}
}

func TestInferSchema_Array(t *testing.T) {
	data := []byte(`[{"id":1},{"id":2}]`)
	s := schemautil.InferSchema(data)
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if s.Type != "array" {
		t.Fatalf("expected type array, got %v", s.Type)
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
	s := schemautil.InferSchema(data)
	prop := s.Properties["count"]
	if prop.Type != "integer" {
		t.Fatalf("expected integer, got %v", prop.Type)
	}
}

func TestInferSchema_Number(t *testing.T) {
	data := []byte(`{"price":9.99}`)
	s := schemautil.InferSchema(data)
	prop := s.Properties["price"]
	if prop.Type != "number" {
		t.Fatalf("expected number, got %v", prop.Type)
	}
}
