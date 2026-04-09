package gswag

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/swaggest/openapi-go/openapi3"
)

// locationToParamIn converts a ParameterLocation to ParameterIn.
// Using WithIn (instead of WithLocation) avoids duplicate "in" keys in marshaled JSON.
func locationToParamIn(loc openapi3.ParameterLocation) openapi3.ParameterIn {
	switch {
	case loc.PathParameter != nil:
		return openapi3.ParameterInPath
	case loc.HeaderParameter != nil:
		return openapi3.ParameterInHeader
	case loc.CookieParameter != nil:
		return openapi3.ParameterInCookie
	default:
		return openapi3.ParameterInQuery
	}
}

// stringParam builds a simple string-typed ParameterOrRef for the given name and location.
func stringParam(name string, loc openapi3.ParameterLocation) openapi3.ParameterOrRef {
	schema := openapi3.Schema{}
	schema.WithType(openapi3.SchemaTypeString)
	sor := openapi3.SchemaOrRef{}
	sor.WithSchema(schema)

	param := openapi3.Parameter{}
	param.WithName(name)
	param.WithIn(locationToParamIn(loc))
	param.WithSchema(sor)

	por := openapi3.ParameterOrRef{}
	por.WithParameter(param)
	return por
}

// dslSchemaParam builds an OpenAPI ParameterOrRef from a DSL dslParam declaration.
func dslSchemaParam(p dslParam, loc openapi3.ParameterLocation) openapi3.ParameterOrRef {
	schemaTypeVal := openapi3.SchemaType(string(p.typ))
	s := openapi3.Schema{}
	s.WithType(schemaTypeVal)
	if p.typ == Array {
		itemSchema := openapi3.Schema{}
		itemSchema.WithType(openapi3.SchemaTypeString)
		itemSor := openapi3.SchemaOrRef{}
		itemSor.WithSchema(itemSchema)
		s.WithItems(itemSor)
	}
	if len(p.enumVals) > 0 {
		s.WithEnum(p.enumVals...)
	}
	if p.hasDef {
		s.WithDefault(p.defVal)
	}
	sor := openapi3.SchemaOrRef{}
	sor.WithSchema(s)

	param := openapi3.Parameter{}
	param.WithName(p.name)
	param.WithIn(locationToParamIn(loc))
	param.WithSchema(sor)
	if p.required != nil {
		param.WithRequired(*p.required)
	}
	if p.explode != nil {
		param.WithExplode(*p.explode)
	}

	por := openapi3.ParameterOrRef{}
	por.WithParameter(param)
	return por
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

// buildPathParamsStruct creates a dynamic struct with `path:"name"` tagged fields
// for each {name} placeholder in pathTemplate. Field type is int64 when the
// concrete value parses as an integer, otherwise string.
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
		fieldName := "P" + string(runes)

		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Type: fieldType,
			Tag:  reflect.StructTag(`path:"` + name + `"`),
		})
	}

	t := reflect.StructOf(fields)
	return reflect.New(t).Interface()
}

// buildPathParamsStructFromDSL creates a dynamic struct for path parameters declared
// via Parameter(name, InPath, schemaType). Undeclared placeholders default to string.
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
		fieldName := "P" + string(runes)

		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Type: fieldType,
			Tag:  reflect.StructTag(`path:"` + name + `"`),
		})
	}

	t := reflect.StructOf(fields)
	return reflect.New(t).Interface()
}

// appendParams adds individual query and header parameters from requestBuilder to
// the already-registered operation.
func (sc *SpecCollector) appendParams(b *requestBuilder) {
	if len(b.queryParams) == 0 && len(b.headers) == 0 {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[b.path]
	if !ok || pathItem.MapOfOperationValues == nil {
		return
	}

	methodKey := strings.ToLower(b.method)
	op, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	for name := range b.queryParams {
		op.Parameters = append(op.Parameters, stringParam(name, openapi3.ParameterLocation{
			QueryParameter: &openapi3.QueryParameter{},
		}))
	}
	for name := range b.headers {
		op.Parameters = append(op.Parameters, stringParam(name, openapi3.ParameterLocation{
			HeaderParameter: &openapi3.HeaderParameter{},
		}))
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// appendDSLParams adds query- and header-typed parameters from DSL Parameter() calls
// to an already-registered operation.
func (sc *SpecCollector) appendDSLParams(op *dslOp) {
	var queryParams []dslParam
	var headerParams []dslParam
	for _, p := range op.params {
		switch p.location {
		case InQuery:
			queryParams = append(queryParams, p)
		case InHeader:
			headerParams = append(headerParams, p)
		case InPath, InCookie:
			// handled elsewhere
		}
	}
	if len(queryParams) == 0 && len(headerParams) == 0 {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[op.path]
	if !ok || pathItem.MapOfOperationValues == nil {
		return
	}
	methodKey := strings.ToLower(op.method)
	operation, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	for _, p := range queryParams {
		operation.Parameters = append(operation.Parameters, dslSchemaParam(p, openapi3.ParameterLocation{
			QueryParameter: &openapi3.QueryParameter{},
		}))
	}
	for _, p := range headerParams {
		operation.Parameters = append(operation.Parameters, dslSchemaParam(p, openapi3.ParameterLocation{
			HeaderParameter: &openapi3.HeaderParameter{},
		}))
	}

	pathItem.MapOfOperationValues[methodKey] = operation
	sc.reflector.Spec.Paths.MapOfPathItemValues[op.path] = pathItem
}
