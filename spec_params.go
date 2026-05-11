package gswag

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/oaswrap/spec/openapi"
)

// locationToParamIn converts a ParamLocation to an OpenAPI parameter location.
func locationToParamIn(loc ParamLocation) string {
	switch loc {
	case InPath:
		return string(openapi.ParameterInPath)
	case InQuery:
		return string(openapi.ParameterInQuery)
	case InHeader:
		return string(openapi.ParameterInHeader)
	case InCookie:
		return string(openapi.ParameterInCookie)
	default:
		return string(openapi.ParameterInQuery)
	}
}

// stringParam builds a simple string-typed parameter.
func stringParam(name string, loc ParamLocation) *openapi.Parameter {
	return &openapi.Parameter{Name: name, In: locationToParamIn(loc), Schema: &openapi.Schema{Type: string(String)}}
}

// dslSchemaParam builds an OpenAPI parameter from a DSL dslParam declaration.
func dslSchemaParam(p dslParam, loc ParamLocation) *openapi.Parameter {
	s := &openapi.Schema{Type: string(p.typ)}
	if p.typ == Array {
		s.Items = &openapi.Schema{Type: string(String)}
	}
	if len(p.enumVals) > 0 {
		s.Enum = append([]any(nil), p.enumVals...)
	}
	if p.hasDef {
		s.Default = p.defVal
	}
	param := &openapi.Parameter{Name: p.name, In: locationToParamIn(loc), Schema: s}
	if p.required != nil {
		param.Required = *p.required
	}
	if p.explode != nil {
		param.Explode = p.explode
	}
	return param
}

// dslSchemaTypeToReflect maps a SchemaType constant to a Go reflect.Type.
func dslSchemaTypeToReflect(typ SchemaType) reflect.Type {
	switch typ {
	case Integer:
		return reflect.TypeFor[int64]()
	case Number:
		return reflect.TypeFor[float64]()
	case Boolean:
		return reflect.TypeFor[bool]()
	case String:
		return reflect.TypeFor[string]()
	case Object:
		return reflect.TypeFor[map[string]any]()
	case Array:
		return reflect.TypeFor[[]string]()
	default:
		return reflect.TypeFor[string]()
	}
}

// buildPathParamsStruct creates a dynamic struct with `path:"name"` tagged fields.
func buildPathParamsStruct(pathTemplate string, pathParamValues map[string]string) any {
	matches := pathParamRe.FindAllStringSubmatch(pathTemplate, -1)
	if len(matches) == 0 {
		return nil
	}
	fields := make([]reflect.StructField, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		fieldType := reflect.TypeFor[string]()
		if val, ok := pathParamValues[name]; ok {
			if _, err := strconv.ParseInt(val, 10, 64); err == nil {
				fieldType = reflect.TypeFor[int64]()
			}
		}
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		fields = append(
			fields,
			reflect.StructField{
				Name: "P" + string(runes),
				Type: fieldType,
				Tag:  reflect.StructTag(`path:"` + name + `"`),
			},
		)
	}
	return reflect.New(reflect.StructOf(fields)).Interface()
}

// buildPathParamsStructFromDSL creates a dynamic struct for path parameters declared via Parameter().
func buildPathParamsStructFromDSL(pathTemplate string, params []dslParam) any {
	matches := pathParamRe.FindAllStringSubmatch(pathTemplate, -1)
	if len(matches) == 0 {
		return nil
	}
	declaredTypes := make(map[string]SchemaType, len(params))
	for _, p := range params {
		if p.location == InPath {
			declaredTypes[p.name] = p.typ
		}
	}
	fields := make([]reflect.StructField, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		fieldType := dslSchemaTypeToReflect(declaredTypes[name])
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		fields = append(
			fields,
			reflect.StructField{
				Name: "P" + string(runes),
				Type: fieldType,
				Tag:  reflect.StructTag(`path:"` + name + `"`),
			},
		)
	}
	return reflect.New(reflect.StructOf(fields)).Interface()
}

// appendParams adds individual query and header parameters from requestBuilder.
func (sc *SpecCollector) appendParams(b *requestBuilder) {
	if len(b.queryParams) == 0 && len(b.headers) == 0 {
		return
	}
	op, ok := sc.operation(b.method, b.path)
	if !ok {
		return
	}
	for name := range b.queryParams {
		op.Parameters = append(op.Parameters, stringParam(name, InQuery))
	}
	for name := range b.headers {
		op.Parameters = append(op.Parameters, stringParam(name, InHeader))
	}
	op.Parameters = sanitizeOperationParameters(b.path, op.Parameters)
}

// appendDSLParams adds query- and header-typed parameters from DSL Parameter() calls.
func (sc *SpecCollector) appendDSLParams(op *dslOp) {
	operation, ok := sc.operation(op.method, op.path)
	if !ok {
		return
	}
	for _, p := range op.params {
		switch p.location {
		case InPath:
			operation.Parameters = append(operation.Parameters, dslSchemaParam(p, InPath))
		case InQuery:
			operation.Parameters = append(operation.Parameters, dslSchemaParam(p, InQuery))
		case InHeader:
			operation.Parameters = append(operation.Parameters, dslSchemaParam(p, InHeader))
		case InCookie:
			operation.Parameters = append(operation.Parameters, dslSchemaParam(p, InCookie))
		}
	}
	operation.Parameters = sanitizeOperationParameters(op.path, operation.Parameters)
}

func sanitizeOperationParameters(path string, params []*openapi.Parameter) []*openapi.Parameter {
	out := make([]*openapi.Parameter, 0, len(params))
	seen := make(map[string]struct{}, len(params))
	for _, p := range params {
		if p == nil || strings.TrimSpace(p.Name) == "" {
			continue
		}
		if strings.TrimSpace(p.In) == "" {
			if strings.Contains(path, "{"+p.Name+"}") {
				p.In = string(openapi.ParameterInPath)
				p.Required = true
			} else {
				p.In = string(openapi.ParameterInQuery)
			}
		}
		key := strings.ToLower(p.In + "|" + p.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}
