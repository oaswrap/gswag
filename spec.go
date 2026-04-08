package gswag

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/oaswrap/gswag/internal/schemautil"
	"github.com/swaggest/jsonschema-go"
	openapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

// bearerAuthSchemeName is the conventional component key for Bearer JWT schemes.
const bearerAuthSchemeName = "bearerAuth"

var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

// SpecCollector accumulates OpenAPI operations from test executions in a thread-safe manner.
type SpecCollector struct {
	mu           sync.Mutex
	reflector    *openapi3.Reflector
	excludePaths []string
}

func newSpecCollector(cfg *Config) *SpecCollector {
	r := openapi3.NewReflector()
	// Apply optional config to underlying JSON Schema reflector.
	if len(cfg.StripDefinitionNamePrefixes) > 0 {
		r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
			jsonschema.StripDefinitionNamePrefix(cfg.StripDefinitionNamePrefixes...))
	}
	if cfg.InlineRefs {
		r.JSONSchemaReflector().DefaultOptions = append(r.JSONSchemaReflector().DefaultOptions,
			jsonschema.InlineRefs)
	}
	if len(cfg.TypeMappings) > 0 {
		for _, m := range cfg.TypeMappings {
			r.JSONSchemaReflector().AddTypeMapping(m.Src, m.Dst)
		}
	}
	r.Spec.Info.
		WithTitle(cfg.Title).
		WithVersion(cfg.Version)

	if cfg.Description != "" {
		r.Spec.Info.WithDescription(cfg.Description)
	}
	if cfg.TermsOfService != "" {
		r.Spec.Info.WithTermsOfService(cfg.TermsOfService)
	}
	if cfg.Contact != nil {
		c := openapi3.Contact{}
		if cfg.Contact.Name != "" {
			c.WithName(cfg.Contact.Name)
		}
		if cfg.Contact.URL != "" {
			c.WithURL(cfg.Contact.URL)
		}
		if cfg.Contact.Email != "" {
			c.WithEmail(cfg.Contact.Email)
		}
		r.Spec.Info.WithContact(c)
	}
	if cfg.License != nil {
		l := openapi3.License{}
		if cfg.License.Name != "" {
			l.WithName(cfg.License.Name)
		}
		if cfg.License.URL != "" {
			l.WithURL(cfg.License.URL)
		}
		r.Spec.Info.WithLicense(l)
	}

	if cfg.ExternalDocs != nil && cfg.ExternalDocs.URL != "" {
		ed := openapi3.ExternalDocumentation{}
		ed.WithURL(cfg.ExternalDocs.URL)
		if cfg.ExternalDocs.Description != "" {
			ed.WithDescription(cfg.ExternalDocs.Description)
		}
		r.Spec.WithExternalDocs(ed)
	}

	if len(cfg.Tags) > 0 {
		tags := make([]openapi3.Tag, 0, len(cfg.Tags))
		for _, tc := range cfg.Tags {
			if tc.Name == "" {
				continue
			}
			t := openapi3.Tag{}
			t.WithName(tc.Name)
			if tc.Description != "" {
				t.WithDescription(tc.Description)
			}
			if tc.ExternalDocs != nil && tc.ExternalDocs.URL != "" {
				ed := openapi3.ExternalDocumentation{}
				ed.WithURL(tc.ExternalDocs.URL)
				if tc.ExternalDocs.Description != "" {
					ed.WithDescription(tc.ExternalDocs.Description)
				}
				t.WithExternalDocs(ed)
			}
			tags = append(tags, t)
		}
		if len(tags) > 0 {
			r.Spec.WithTags(tags...)
		}
	}

	for _, srv := range cfg.Servers {
		s := openapi3.Server{}
		s.WithURL(srv.URL)
		if srv.Description != "" {
			s.WithDescription(srv.Description)
		}
		r.Spec.WithServers(s)
	}

	sc := &SpecCollector{reflector: r, excludePaths: append([]string(nil), cfg.ExcludePaths...)}

	// Pre-register security schemes declared in Config.
	for name, schemeCfg := range cfg.SecuritySchemes {
		sc.reflector.Spec.ComponentsEns().SecuritySchemesEns().
			WithMapOfSecuritySchemeOrRefValuesItem(name, buildSecuritySchemeOrRef(schemeCfg))
	}

	return sc
}

// Register adds an operation to the spec based on the requestBuilder metadata
// and the actual recordedResponse. Safe to call concurrently.
func (sc *SpecCollector) Register(b *requestBuilder, res *recordedResponse) {
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
	if sc.isExcludedPath(b.path) {
		return
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

	// Inject request body schema/media type from the actual request when needed.
	sc.injectInferredRequestSchema(b, res)

	// Inject inferred JSON schema into the fallback response slot.
	if len(b.respBodies) == 0 {
		sc.injectInferredSchema(b, res)
	}

	// Append individual query and header parameters collected via WithQueryParam / WithHeader.
	sc.appendParams(b)

	// Append declared response header schemas.
	sc.appendResponseHeaders(b)

	// Append captured examples for request/response bodies (if enabled).
	sc.appendExamplesLocked(b, res)
}

// injectInferredRequestSchema attaches request body media type/schema from the
// actual request when a request body exists at runtime.
func (sc *SpecCollector) injectInferredRequestSchema(b *requestBuilder, res *recordedResponse) {
	if len(res.RequestBodyBytes) == 0 {
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

	ct := strings.TrimSpace(b.bodyContentType)
	// Strip content-type parameters (e.g. boundary= from multipart/form-data)
	// so the media type key in the spec is the canonical base type only.
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" {
		if b.body != nil {
			ct = "application/json"
		} else {
			ct = "application/octet-stream"
		}
	}

	var schema *openapi3.SchemaOrRef
	if strings.Contains(strings.ToLower(ct), "json") {
		schema = schemautil.InferSchema(res.RequestBodyBytes)
	}
	if schema == nil {
		s := openapi3.Schema{}
		s.WithType(openapi3.SchemaTypeString)
		s.WithFormat("binary")
		sor := openapi3.SchemaOrRef{}
		sor.WithSchema(s)
		schema = &sor
	}

	or := op.RequestBodyEns()
	rb := or.RequestBodyEns()
	if rb.Content == nil {
		rb.Content = map[string]openapi3.MediaType{}
	}

	// If any content-type key already has a schema (e.g. placed by Consumes +
	// RequestBody via the DSL), do not inject an inferred schema under a
	// potentially different key from SetRawBody. That would produce conflicting
	// entries in the spec.
	for _, existingMT := range rb.Content {
		if existingMT.Schema != nil {
			pathItem.MapOfOperationValues[methodKey] = op
			sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
			return
		}
	}

	mt := rb.Content[ct]
	if mt.Schema == nil {
		mt.Schema = schema
		rb.Content[ct] = mt
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// injectInferredSchema parses the response body and attaches a best-effort
// OpenAPI schema to the already-registered response.
func (sc *SpecCollector) injectInferredSchema(b *requestBuilder, res *recordedResponse) {
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
func (sc *SpecCollector) appendParams(b *requestBuilder) {
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

// appendResponseHeaders attaches any response header schemas declared via the
// requestBuilder to the corresponding response objects in the registered spec.
func (sc *SpecCollector) appendResponseHeaders(b *requestBuilder) {
	if len(b.respHeaders) == 0 {
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

	for status, headers := range b.respHeaders {
		statusKey := strconv.Itoa(status)
		respOrRef, ok := op.Responses.MapOfResponseOrRefValues[statusKey]
		if !ok || respOrRef.Response == nil {
			continue
		}
		resp := respOrRef.Response
		if resp.Headers == nil {
			resp.Headers = make(map[string]openapi3.HeaderOrRef)
		}

		for name, model := range headers {
			// Marshal model to JSON and infer a schema OR use inference from example bytes.
			var sor *openapi3.SchemaOrRef
			if model == nil {
				// default to string
				schema := openapi3.Schema{}
				schema.WithType(openapi3.SchemaTypeString)
				s := openapi3.SchemaOrRef{}
				s.WithSchema(schema)
				sor = &s
			} else {
				bts, err := json.Marshal(model)
				if err != nil {
					// fallback to string schema on marshal error
					schema := openapi3.Schema{}
					schema.WithType(openapi3.SchemaTypeString)
					s := openapi3.SchemaOrRef{}
					s.WithSchema(schema)
					sor = &s
				} else {
					inferred := schemautil.InferSchema(bts)
					if inferred == nil {
						schema := openapi3.Schema{}
						schema.WithType(openapi3.SchemaTypeString)
						s := openapi3.SchemaOrRef{}
						s.WithSchema(schema)
						sor = &s
					} else {
						sor = inferred
					}
				}
			}

			header := openapi3.Header{}
			header.WithSchema(*sor)
			har := openapi3.HeaderOrRef{}
			har.WithHeader(header)
			resp.Headers[name] = har
		}

		respOrRef.Response = resp
		op.Responses.MapOfResponseOrRefValues[statusKey] = respOrRef
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// locationToParamIn converts a ParameterLocation to the corresponding ParameterIn
// value. Using WithIn (instead of WithLocation) avoids duplicate "in" keys in the
// marshaled JSON, which occurs because Parameter.MarshalJSON merges both
// marshalParameter (emitting p.In) and p.Location (emitting its own const "in").
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

func stringParam(name string, loc openapi3.ParameterLocation) openapi3.ParameterOrRef {
	schemaType := openapi3.SchemaTypeString
	schema := openapi3.Schema{}
	schema.WithType(schemaType)
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

// appendExamples attaches captured request/response examples to the spec
// if the global config enables example capture.
func (sc *SpecCollector) appendExamples(b *requestBuilder, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.appendExamplesLocked(b, res)
}

func (sc *SpecCollector) appendExamplesLocked(b *requestBuilder, res *recordedResponse) {
	if globalConfig == nil || !globalConfig.CaptureExamples {
		return
	}

	// helper to apply sanitizer and cap
	sanitize := func(in []byte) []byte {
		if in == nil {
			return nil
		}
		out := in
		if globalConfig.Sanitizer != nil {
			out = globalConfig.Sanitizer(out)
		}
		if globalConfig.MaxExampleBytes > 0 && len(out) > globalConfig.MaxExampleBytes {
			out = out[:globalConfig.MaxExampleBytes]
		}
		return out
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

	// Request body example
	if op.RequestBody != nil && op.RequestBody.RequestBody != nil && len(res.RequestBodyBytes) > 0 {
		rb := op.RequestBody.RequestBody
		ct := requestExampleContentType(b, res)
		if rb.Content != nil {
			if mt, found := rb.Content[ct]; found {
				bts := sanitize(res.RequestBodyBytes)
				var ex interface{}
				if err := json.Unmarshal(bts, &ex); err != nil {
					ex = string(bts)
				}
				mt.Example = &ex
				rb.Content[ct] = mt
				op.RequestBody.RequestBody = rb
			}
		}
	}

	// Response body example for the actual status code
	statusKey := strconv.Itoa(res.StatusCode)
	if op.Responses.MapOfResponseOrRefValues != nil {
		if ror, found := op.Responses.MapOfResponseOrRefValues[statusKey]; found && ror.Response != nil {
			resp := ror.Response
			ct := responseExampleContentType(res)
			if resp.Content != nil {
				if mt, found := resp.Content[ct]; found {
					bts := sanitize(res.BodyBytes)
					var ex interface{}
					if err := json.Unmarshal(bts, &ex); err != nil {
						ex = string(bts)
					}
					mt.Example = &ex
					resp.Content[ct] = mt
					ror.Response = resp
					op.Responses.MapOfResponseOrRefValues[statusKey] = ror
				}
			}
		}
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

func requestExampleContentType(b *requestBuilder, res *recordedResponse) string {
	if b != nil {
		if ct := normalizeContentType(b.bodyContentType); ct != "" {
			return ct
		}
		if b.body != nil || len(res.RequestBodyBytes) > 0 {
			return "application/json"
		}
	}
	return "application/json"
}

func responseExampleContentType(res *recordedResponse) string {
	if res != nil && res.Headers != nil {
		if ct := normalizeContentType(res.Headers.Get("Content-Type")); ct != "" {
			return ct
		}
	}
	return "application/json"
}

func normalizeContentType(ct string) string {
	ct = strings.TrimSpace(ct)
	if ct == "" {
		return ""
	}
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = ct[:idx]
	}
	return strings.TrimSpace(ct)
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

	case "oauth2":
		flows := openapi3.OAuthFlows{}
		implicit := openapi3.ImplicitOAuthFlow{}
		implicit.WithAuthorizationURL(cfg.AuthorizationURL)
		if cfg.RefreshURL != "" {
			implicit.WithRefreshURL(cfg.RefreshURL)
		}
		if len(cfg.Scopes) > 0 {
			implicit.WithScopes(cfg.Scopes)
		} else {
			implicit.WithScopes(map[string]string{})
		}
		flows.WithImplicit(implicit)

		oo := openapi3.OAuth2SecurityScheme{}
		oo.WithFlows(flows)
		scheme.WithOAuth2SecurityScheme(oo)

	case "openidconnect":
		oid := openapi3.OpenIDConnectSecurityScheme{}
		oid.WithOpenIDConnectURL(cfg.AuthorizationURL)
		scheme.WithOpenIDConnectSecurityScheme(oid)
	}

	sor.WithSecurityScheme(scheme)
	return sor
}

// RegisterDSLOperation registers an operation declared via the rswag-style DSL.
// It is called from a Ginkgo BeforeAll node so that spec registration happens
// once per operation, before any RunTest It blocks execute.
func (sc *SpecCollector) RegisterDSLOperation(op *dslOp) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if op.hidden || sc.isExcludedPath(op.path) {
		return
	}

	opCtx, err := sc.reflector.NewOperationContext(op.method, op.path)
	if err != nil {
		fmt.Printf("gswag DSL: NewOperationContext error for %s %s: %v\n", op.method, op.path, err)
		return
	}

	if len(op.tags) > 0 {
		opCtx.SetTags(op.tags...)
	}
	if op.summary != "" {
		opCtx.SetSummary(op.summary)
	}
	if op.description != "" {
		opCtx.SetDescription(op.description)
	}
	if op.operationID != "" {
		opCtx.SetID(op.operationID)
	}
	if op.deprecated {
		opCtx.SetIsDeprecated(true)
	}
	for _, sec := range op.security {
		for name, scopes := range sec {
			opCtx.AddSecurity(name, scopes...)
		}
	}

	// Path parameters from declared Parameter() calls.
	if pathStruct := buildPathParamsStructFromDSL(op.path, op.params); pathStruct != nil {
		opCtx.AddReqStructure(pathStruct)
	}

	// Typed query param struct.
	if op.queryStruct != nil {
		opCtx.AddReqStructure(op.queryStruct)
	}

	// Request body schema.
	if op.reqBodyModel != nil {
		if op.consumes != "" {
			opCtx.AddReqStructure(op.reqBodyModel, openapi.WithContentType(op.consumes))
		} else {
			opCtx.AddReqStructure(op.reqBodyModel)
		}
	}

	// Response schemas — one entry per declared Response() block.
	if len(op.responses) > 0 {
		for status, resp := range op.responses {
			s := status
			var model interface{}
			if resp != nil {
				model = resp.bodyModel
			}
			if len(op.produces) > 0 {
				for _, ct := range op.produces {
					contentType := ct
					opCtx.AddRespStructure(model, openapi.WithContentType(contentType), func(cu *openapi.ContentUnit) {
						cu.HTTPStatus = s
					})
				}
			} else {
				opCtx.AddRespStructure(model, func(cu *openapi.ContentUnit) {
					cu.HTTPStatus = s
				})
			}
		}
	} else {
		opCtx.AddRespStructure(nil, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = 200
		})
	}

	// Ensure all referenced security schemes are declared in components.
	for _, sec := range op.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	if err := sc.reflector.AddOperation(opCtx); err != nil {
		fmt.Printf("gswag DSL: AddOperation error for %s %s: %v\n", op.method, op.path, err)
		return
	}

	// Append individual query/header parameters (param location != path).
	sc.appendDSLParams(op)

	// Append declared response header schemas.
	sc.appendDSLResponseHeaders(op)
}

func (sc *SpecCollector) isExcludedPath(path string) bool {
	if len(sc.excludePaths) == 0 {
		return false
	}

	for _, pattern := range sc.excludePaths {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if pattern == path {
			return true
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	return false
}

// injectRecordedResponseSchema injects an inferred schema from the actual response body
// into an existing operation response slot that has no explicit schema declared.
// Called from RunTest after the HTTP request fires.
func (sc *SpecCollector) injectRecordedResponseSchema(method, path string, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[path]
	if !ok {
		return
	}
	if pathItem.MapOfOperationValues == nil {
		return
	}
	methodKey := strings.ToLower(method)
	op, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	statusKey := strconv.Itoa(res.StatusCode)
	if op.Responses.MapOfResponseOrRefValues == nil {
		return
	}
	ror, ok := op.Responses.MapOfResponseOrRefValues[statusKey]
	if !ok || ror.Response == nil {
		return
	}

	// Skip if a schema is already present.
	resp := ror.Response
	ct := "application/json"
	if resp.Content != nil {
		if mt, found := resp.Content[ct]; found && mt.Schema != nil {
			return
		}
	}

	// Infer from the actual body bytes.
	inferred := schemautil.InferSchema(res.BodyBytes)
	if inferred == nil {
		return
	}

	if resp.Content == nil {
		resp.Content = map[string]openapi3.MediaType{
			ct: {Schema: inferred},
		}
	} else {
		mt := resp.Content[ct]
		if mt.Schema == nil {
			mt.Schema = inferred
			resp.Content[ct] = mt
		}
	}

	ror.Response = resp
	op.Responses.MapOfResponseOrRefValues[statusKey] = ror
	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[path] = pathItem
}

// appendDSLParams adds query- and header-typed parameters (from Parameter() DSL calls)
// to an already-registered operation in the spec.
func (sc *SpecCollector) appendDSLParams(op *dslOp) {
	var queryParams []dslParam
	var headerParams []dslParam
	for _, p := range op.params {
		switch p.location {
		case InQuery:
			queryParams = append(queryParams, p)
		case InHeader:
			headerParams = append(headerParams, p)
		}
	}
	if len(queryParams) == 0 && len(headerParams) == 0 {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[op.path]
	if !ok {
		return
	}
	if pathItem.MapOfOperationValues == nil {
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

// appendDSLResponseHeaders attaches response header schemas declared via ResponseHeader()
// to the corresponding response objects for the operation.
func (sc *SpecCollector) appendDSLResponseHeaders(op *dslOp) {
	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[op.path]
	if !ok {
		return
	}
	if pathItem.MapOfOperationValues == nil {
		return
	}
	methodKey := strings.ToLower(op.method)
	operation, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	for status, resp := range op.responses {
		if resp == nil || len(resp.headers) == 0 {
			continue
		}
		statusKey := strconv.Itoa(status)
		respOrRef, found := operation.Responses.MapOfResponseOrRefValues[statusKey]
		if !found || respOrRef.Response == nil {
			continue
		}
		r := respOrRef.Response
		if r.Headers == nil {
			r.Headers = make(map[string]openapi3.HeaderOrRef)
		}
		for name, model := range resp.headers {
			var sor *openapi3.SchemaOrRef
			if model == nil {
				s := openapi3.Schema{}
				s.WithType(openapi3.SchemaTypeString)
				so := openapi3.SchemaOrRef{}
				so.WithSchema(s)
				sor = &so
			} else {
				bts, err := json.Marshal(model)
				if err == nil {
					sor = schemautil.InferSchema(bts)
				}
				if sor == nil {
					s := openapi3.Schema{}
					s.WithType(openapi3.SchemaTypeString)
					so := openapi3.SchemaOrRef{}
					so.WithSchema(s)
					sor = &so
				}
			}
			h := openapi3.Header{}
			h.WithSchema(*sor)
			har := openapi3.HeaderOrRef{}
			har.WithHeader(h)
			r.Headers[name] = har
		}
		respOrRef.Response = r
		operation.Responses.MapOfResponseOrRefValues[statusKey] = respOrRef
	}

	pathItem.MapOfOperationValues[methodKey] = operation
	sc.reflector.Spec.Paths.MapOfPathItemValues[op.path] = pathItem
}

// buildPathParamsStructFromDSL creates a dynamic struct for the path parameters
// declared via Parameter(name, InPath, schemaType). Falls back to string for any
// path placeholder not explicitly declared.
func buildPathParamsStructFromDSL(pathTemplate string, params []dslParam) interface{} {
	matches := pathParamRe.FindAllStringSubmatch(pathTemplate, -1)
	if len(matches) == 0 {
		return nil
	}

	// Build a lookup from param name → declared schema type.
	declaredTypes := make(map[string]SchemaType, len(params))
	for _, p := range params {
		if p.location == InPath {
			declaredTypes[p.name] = p.typ
		}
	}

	fields := make([]reflect.StructField, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		fieldType := dslSchemaTypeToReflect(declaredTypes[name]) // defaults to string when not declared

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

// dslSchemaTypeToReflect maps a SchemaType to a Go reflect.Type for struct-field generation.
func dslSchemaTypeToReflect(typ SchemaType) reflect.Type {
	switch typ {
	case Integer:
		return reflect.TypeOf(int64(0))
	case Number:
		return reflect.TypeOf(float64(0))
	case Boolean:
		return reflect.TypeOf(false)
	default:
		return reflect.TypeOf("")
	}
}

// dslSchemaParam builds an OpenAPI ParameterOrRef with the given name, schema type, and location.
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
