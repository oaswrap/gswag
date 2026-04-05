package gswag

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	openapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/oaswrap/gswag/internal/schemautil"
)

// bearerAuthSchemeName is the conventional component key for Bearer JWT schemes.
const bearerAuthSchemeName = "bearerAuth"

var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

// SpecCollector accumulates OpenAPI operations from test executions in a thread-safe manner.
type SpecCollector struct {
	mu        sync.Mutex
	reflector *openapi3.Reflector
}

func newSpecCollector(cfg *Config) *SpecCollector {
	r := openapi3.Reflector{}
	r.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	r.Spec.Info.
		WithTitle(cfg.Title).
		WithVersion(cfg.Version)

	if cfg.Description != "" {
		r.Spec.Info.WithDescription(cfg.Description)
	}

	for _, srv := range cfg.Servers {
		s := openapi3.Server{}
		s.WithURL(srv.URL)
		if srv.Description != "" {
			s.WithDescription(srv.Description)
		}
		r.Spec.WithServers(s)
	}

	sc := &SpecCollector{reflector: &r}

	// Pre-register security schemes declared in Config.
	for name, schemeCfg := range cfg.SecuritySchemes {
		sc.reflector.Spec.ComponentsEns().SecuritySchemesEns().
			WithMapOfSecuritySchemeOrRefValuesItem(name, buildSecuritySchemeOrRef(schemeCfg))
	}

	return sc
}

// Register adds an operation to the spec based on the RequestBuilder metadata
// and the actual RecordedResponse. Safe to call concurrently.
func (sc *SpecCollector) Register(b *RequestBuilder, res *RecordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	op, err := sc.reflector.NewOperationContext(b.method, b.path)
	if err != nil {
		fmt.Printf("gswag: NewOperationContext error for %s %s: %v\n", b.method, b.path, err)
		return
	}

	if len(b.tags) > 0 {
		op.SetTags(b.tags...)
	}
	if b.summary != "" {
		op.SetSummary(b.summary)
	}
	if b.description != "" {
		op.SetDescription(b.description)
	}
	if b.operationID != "" {
		op.SetID(b.operationID)
	}
	if b.deprecated {
		op.SetIsDeprecated(true)
	}
	for _, sec := range b.security {
		for name, scopes := range sec {
			op.AddSecurity(name, scopes...)
		}
	}

	// Path parameters — must be declared before AddOperation to pass the reflector's validation.
	if pathStruct := buildPathParamsStruct(b.path, b.pathParams); pathStruct != nil {
		op.AddReqStructure(pathStruct)
	}

	// Typed query param struct (fields with `query` tags).
	if b.queryStruct != nil {
		op.AddReqStructure(b.queryStruct)
	}

	// Request body (typed struct → JSON schema).
	if b.body != nil {
		op.AddReqStructure(b.body)
	}

	// Response schemas.
	if len(b.respBodies) > 0 {
		for status, model := range b.respBodies {
			s := status
			op.AddRespStructure(model, func(cu *openapi.ContentUnit) {
				cu.HTTPStatus = s
			})
		}
	} else {
		// Fallback: emit an empty response for the actual status code.
		// The inferred schema is attached after AddOperation below.
		status := res.StatusCode
		op.AddRespStructure(nil, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = status
		})
	}

	// Ensure every security scheme referenced by the operation is declared in components.
	for _, sec := range b.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	if err := sc.reflector.AddOperation(op); err != nil {
		fmt.Printf("gswag: AddOperation error for %s %s: %v\n", b.method, b.path, err)
		return
	}

	// Inject inferred JSON schema into the fallback response slot.
	if len(b.respBodies) == 0 {
		sc.injectInferredSchema(b, res)
	}

	// Append individual query and header parameters collected via WithQueryParam / WithHeader.
	sc.appendParams(b)
}

// injectInferredSchema parses the response body and attaches a best-effort
// OpenAPI schema to the already-registered response.
func (sc *SpecCollector) injectInferredSchema(b *RequestBuilder, res *RecordedResponse) {
	inferred := schemautil.InferSchema(res.BodyBytes)
	if inferred == nil {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[b.path]
	if !ok {
		return
	}
	if pathItem.MapOfOperationValues == nil {
		return
	}

	methodKey := strings.ToLower(b.method)
	operation, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	statusKey := strconv.Itoa(res.StatusCode)
	if operation.Responses.MapOfResponseOrRefValues == nil {
		return
	}

	respOrRef, ok := operation.Responses.MapOfResponseOrRefValues[statusKey]
	if !ok || respOrRef.Response == nil {
		return
	}

	resp := respOrRef.Response
	if resp.Content == nil {
		ct := "application/json"
		resp.Content = map[string]openapi3.MediaType{
			ct: {Schema: inferred},
		}
	} else {
		ct := "application/json"
		mt := resp.Content[ct]
		if mt.Schema == nil {
			mt.Schema = inferred
			resp.Content[ct] = mt
		}
	}

	respOrRef.Response = resp
	operation.Responses.MapOfResponseOrRefValues[statusKey] = respOrRef
	pathItem.MapOfOperationValues[methodKey] = operation
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// appendParams manually adds individual query and header parameters (set via
// WithQueryParam / WithHeader) to the already-registered operation.
func (sc *SpecCollector) appendParams(b *RequestBuilder) {
	if len(b.queryParams) == 0 && len(b.headers) == 0 {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[b.path]
	if !ok {
		return
	}
	if pathItem.MapOfOperationValues == nil {
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

func stringParam(name string, loc openapi3.ParameterLocation) openapi3.ParameterOrRef {
	schemaType := openapi3.SchemaTypeString
	schema := openapi3.Schema{}
	schema.WithType(schemaType)
	sor := openapi3.SchemaOrRef{}
	sor.WithSchema(schema)

	param := openapi3.Parameter{}
	param.WithName(name)
	param.WithLocation(loc)
	param.WithSchema(sor)

	por := openapi3.ParameterOrRef{}
	por.WithParameter(param)
	return por
}

// buildPathParamsStruct creates a dynamic struct with typed fields tagged as
// `path:"name"` for each {name} placeholder found in pathTemplate.
// When a value is provided in pathParamValues and looks like an integer, the field type is int64.
func buildPathParamsStruct(pathTemplate string, pathParamValues map[string]string) interface{} {
	matches := pathParamRe.FindAllStringSubmatch(pathTemplate, -1)
	if len(matches) == 0 {
		return nil
	}

	fields := make([]reflect.StructField, 0, len(matches))
	for _, m := range matches {
		name := m[1]

		// Infer field type from the concrete value when available.
		fieldType := reflect.TypeOf("")
		if val, ok := pathParamValues[name]; ok {
			if _, err := strconv.ParseInt(val, 10, 64); err == nil {
				fieldType = reflect.TypeOf(int64(0))
			}
		}

		// Field name must be exported.
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

// ensureSecurityScheme checks whether name is already declared in components/securitySchemes.
// If not, it auto-registers well-known built-in schemes (currently "bearerAuth").
// Unknown names are silently ignored — callers should use Config.SecuritySchemes.
// Must be called with sc.mu held.
func (sc *SpecCollector) ensureSecurityScheme(name string) {
	schemes := sc.reflector.Spec.ComponentsEns().SecuritySchemesEns()
	if _, exists := schemes.MapOfSecuritySchemeOrRefValues[name]; exists {
		return
	}

	switch name {
	case bearerAuthSchemeName:
		http := openapi3.HTTPSecurityScheme{}
		http.WithScheme("bearer").WithBearerFormat("JWT")
		scheme := openapi3.SecurityScheme{}
		scheme.WithHTTPSecurityScheme(http)
		sor := openapi3.SecuritySchemeOrRef{}
		sor.WithSecurityScheme(scheme)
		schemes.WithMapOfSecuritySchemeOrRefValuesItem(name, sor)
	}
}

// buildSecuritySchemeOrRef converts a SecuritySchemeConfig into its openapi3 representation.
func buildSecuritySchemeOrRef(cfg SecuritySchemeConfig) openapi3.SecuritySchemeOrRef {
	sor := openapi3.SecuritySchemeOrRef{}
	scheme := openapi3.SecurityScheme{}

	switch strings.ToLower(cfg.Type) {
	case "http":
		h := openapi3.HTTPSecurityScheme{}
		h.WithScheme(cfg.Scheme)
		if cfg.BearerFormat != "" {
			h.WithBearerFormat(cfg.BearerFormat)
		}
		scheme.WithHTTPSecurityScheme(h)

	case "apikey":
		ak := openapi3.APIKeySecurityScheme{}
		ak.WithName(cfg.Name)
		switch strings.ToLower(cfg.In) {
		case "header":
			ak.WithIn(openapi3.APIKeySecuritySchemeInHeader)
		case "query":
			ak.WithIn(openapi3.APIKeySecuritySchemeInQuery)
		case "cookie":
			ak.WithIn(openapi3.APIKeySecuritySchemeInCookie)
		}
		scheme.WithAPIKeySecurityScheme(ak)
	}

	sor.WithSecurityScheme(scheme)
	return sor
}
