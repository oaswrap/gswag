// Package schemautil provides best-effort JSON → OpenAPI schema inference.
package schemautil

import (
	"encoding/json"
	"math"

	"github.com/swaggest/openapi-go/openapi3"
)

// InferSchema parses raw JSON bytes and returns a best-effort OpenAPI 3.0 schema.
// It returns nil when data is empty or cannot be parsed.
func InferSchema(data []byte) *openapi3.SchemaOrRef {
	if len(data) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil
	}
	s := inferValue(v)
	if s == nil {
		return nil
	}
	sor := &openapi3.SchemaOrRef{}
	sor.WithSchema(*s)
	return sor
}

func inferValue(v interface{}) *openapi3.Schema {
	switch val := v.(type) {
	case map[string]interface{}:
		return inferObject(val)
	case []interface{}:
		return inferArray(val)
	case string:
		t := openapi3.SchemaTypeString
		return (&openapi3.Schema{}).WithType(t)
	case float64:
		if val == math.Trunc(val) {
			t := openapi3.SchemaTypeInteger
			return (&openapi3.Schema{}).WithType(t)
		}
		t := openapi3.SchemaTypeNumber
		return (&openapi3.Schema{}).WithType(t)
	case bool:
		t := openapi3.SchemaTypeBoolean
		return (&openapi3.Schema{}).WithType(t)
	default:
		// null or unknown — return no schema
		return nil
	}
}

func inferObject(m map[string]interface{}) *openapi3.Schema {
	t := openapi3.SchemaTypeObject
	s := (&openapi3.Schema{}).WithType(t)

	if len(m) == 0 {
		return s
	}

	props := make(map[string]openapi3.SchemaOrRef, len(m))
	for k, v := range m {
		child := inferValue(v)
		if child == nil {
			child = &openapi3.Schema{}
		}
		sor := openapi3.SchemaOrRef{}
		sor.WithSchema(*child)
		props[k] = sor
	}
	s.WithProperties(props)
	return s
}

func inferArray(arr []interface{}) *openapi3.Schema {
	t := openapi3.SchemaTypeArray
	s := (&openapi3.Schema{}).WithType(t)

	if len(arr) == 0 {
		return s
	}

	// Infer items schema from the first element.
	child := inferValue(arr[0])
	if child != nil {
		sor := openapi3.SchemaOrRef{}
		sor.WithSchema(*child)
		s.WithItems(sor)
	}

	return s
}
