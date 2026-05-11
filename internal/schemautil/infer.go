// Package schemautil provides best-effort JSON -> OpenAPI schema inference.
package schemautil

import (
	"encoding/json"
	"math"

	"github.com/oaswrap/spec/openapi"
)

// InferSchema parses raw JSON bytes and returns a best-effort OpenAPI schema.
// It returns nil when data is empty or cannot be parsed.
func InferSchema(data []byte) *openapi.Schema {
	if len(data) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil
	}
	return inferValue(v)
}

func inferValue(v any) *openapi.Schema {
	switch val := v.(type) {
	case map[string]any:
		return inferObject(val)
	case []any:
		return inferArray(val)
	case string:
		return &openapi.Schema{Type: "string"}
	case float64:
		if val == math.Trunc(val) {
			return &openapi.Schema{Type: "integer"}
		}
		return &openapi.Schema{Type: "number"}
	case bool:
		return &openapi.Schema{Type: "boolean"}
	default:
		return nil
	}
}

func inferObject(m map[string]any) *openapi.Schema {
	s := &openapi.Schema{Type: "object"}
	if len(m) == 0 {
		return s
	}
	s.Properties = make(map[string]*openapi.Schema, len(m))
	for k, v := range m {
		child := inferValue(v)
		if child == nil {
			child = &openapi.Schema{}
		}
		s.Properties[k] = child
	}
	return s
}

func inferArray(arr []any) *openapi.Schema {
	s := &openapi.Schema{Type: "array"}
	if len(arr) == 0 {
		return s
	}
	if child := inferValue(arr[0]); child != nil {
		s.Items = child
	}
	return s
}
