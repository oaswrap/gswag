package gswag

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/oaswrap/gswag/internal/schemautil"
	"github.com/swaggest/openapi-go/openapi3"
)

// injectInferredRequestSchema attaches an inferred request-body schema from the
// actual request bytes when no explicit schema has been declared.
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

	ct := normalizeContentType(b.bodyContentType)
	if ct == "" {
		if b.body != nil {
			ct = applicationJSON
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

	// If any existing content-type key already has a schema, skip injection to
	// avoid conflicting entries (e.g., from Consumes + RequestBody DSL).
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

// injectInferredSchema parses the response body and attaches a best-effort schema
// to the already-registered fallback response slot.
func (sc *SpecCollector) injectInferredSchema(b *requestBuilder, res *recordedResponse) {
	inferred := schemautil.InferSchema(res.BodyBytes)
	if inferred == nil {
		return
	}

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[b.path]
	if !ok || pathItem.MapOfOperationValues == nil {
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
	ct := applicationJSON
	if resp.Content == nil {
		resp.Content = map[string]openapi3.MediaType{ct: {Schema: inferred}}
	} else {
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

// injectRecordedResponseSchema injects an inferred schema from the actual response body
// into an existing operation response slot that has no explicit schema.
// Called from RunTest after the HTTP request fires.
func (sc *SpecCollector) injectRecordedResponseSchema(method, path string, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[path]
	if !ok || pathItem.MapOfOperationValues == nil {
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

	resp := ror.Response
	ct := applicationJSON
	if resp.Content != nil {
		if mt, found := resp.Content[ct]; found && mt.Schema != nil {
			return
		}
	}

	inferred := schemautil.InferSchema(res.BodyBytes)
	if inferred == nil {
		return
	}

	if resp.Content == nil {
		resp.Content = map[string]openapi3.MediaType{ct: {Schema: inferred}}
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

// appendResponseHeaders attaches declared response header schemas from requestBuilder.
func (sc *SpecCollector) appendResponseHeaders(b *requestBuilder) {
	if len(b.respHeaders) == 0 {
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
			resp.Headers[name] = buildHeaderOrRef(model)
		}
		respOrRef.Response = resp
		op.Responses.MapOfResponseOrRefValues[statusKey] = respOrRef
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// appendDSLResponseHeaders attaches response header schemas declared via ResponseHeader()
// for a DSL operation.
func (sc *SpecCollector) appendDSLResponseHeaders(op *dslOp) {
	pathItem, ok := sc.reflector.Spec.Paths.MapOfPathItemValues[op.path]
	if !ok || pathItem.MapOfOperationValues == nil {
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
			r.Headers[name] = buildHeaderOrRef(model)
		}
		respOrRef.Response = r
		operation.Responses.MapOfResponseOrRefValues[statusKey] = respOrRef
	}

	pathItem.MapOfOperationValues[methodKey] = operation
	sc.reflector.Spec.Paths.MapOfPathItemValues[op.path] = pathItem
}

// buildHeaderOrRef converts a model value to an openapi3.HeaderOrRef with an inferred schema.
func buildHeaderOrRef(model any) openapi3.HeaderOrRef {
	sor := inferHeaderSchema(model)
	h := openapi3.Header{}
	h.WithSchema(*sor)
	har := openapi3.HeaderOrRef{}
	har.WithHeader(h)
	return har
}

func inferHeaderSchema(model any) *openapi3.SchemaOrRef {
	if model == nil {
		return stringSchemaOrRef()
	}
	bts, err := json.Marshal(model)
	if err != nil {
		return stringSchemaOrRef()
	}
	inferred := schemautil.InferSchema(bts)
	if inferred == nil {
		return stringSchemaOrRef()
	}
	return inferred
}

func stringSchemaOrRef() *openapi3.SchemaOrRef {
	s := openapi3.Schema{}
	s.WithType(openapi3.SchemaTypeString)
	so := openapi3.SchemaOrRef{}
	so.WithSchema(s)
	return &so
}

// appendExamples attaches captured request/response examples to the spec (public entry point).
func (sc *SpecCollector) appendExamples(b *requestBuilder, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.appendExamplesLocked(b, res)
}

func (sc *SpecCollector) appendExamplesLocked(b *requestBuilder, res *recordedResponse) {
	if globalConfig == nil || !globalConfig.CaptureExamples {
		return
	}

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
	if !ok || pathItem.MapOfOperationValues == nil {
		return
	}

	methodKey := strings.ToLower(b.method)
	op, ok := pathItem.MapOfOperationValues[methodKey]
	if !ok {
		return
	}

	// Request body example — only captured for successful responses so that
	// error-triggering payloads (e.g. invalid JSON) are not used as examples.
	if res.StatusCode/100 == 2 && op.RequestBody != nil && op.RequestBody.RequestBody != nil && len(res.RequestBodyBytes) > 0 {
		rb := op.RequestBody.RequestBody
		ct := requestExampleContentType(b, res)
		if rb.Content != nil {
			if mt, found := rb.Content[ct]; found {
				bts := sanitize(res.RequestBodyBytes)
				var ex any
				if err := json.Unmarshal(bts, &ex); err == nil {
					mt.Example = &ex
					rb.Content[ct] = mt
					op.RequestBody.RequestBody = rb
				}
			}
		}
	}

	// Response body example for the actual status code.
	statusKey := strconv.Itoa(res.StatusCode)
	if op.Responses.MapOfResponseOrRefValues != nil {
		if ror, found := op.Responses.MapOfResponseOrRefValues[statusKey]; found && ror.Response != nil {
			resp := ror.Response
			ct := responseExampleContentType(res)
			if resp.Content != nil {
				if mt, found := resp.Content[ct]; found {
					bts := sanitize(res.BodyBytes)
					var ex any
					if err := json.Unmarshal(bts, &ex); err == nil {
						mt.Example = &ex
						resp.Content[ct] = mt
						ror.Response = resp
						op.Responses.MapOfResponseOrRefValues[statusKey] = ror
					}
				}
			}
		}
	}

	pathItem.MapOfOperationValues[methodKey] = op
	sc.reflector.Spec.Paths.MapOfPathItemValues[b.path] = pathItem
}

// requestExampleContentType returns the content-type key to use for request examples.
func requestExampleContentType(b *requestBuilder, res *recordedResponse) string {
	if b != nil {
		if ct := normalizeContentType(b.bodyContentType); ct != "" {
			return ct
		}
		if b.body != nil || len(res.RequestBodyBytes) > 0 {
			return applicationJSON
		}
	}
	return applicationJSON
}

// responseExampleContentType returns the content-type key to use for response examples.
func responseExampleContentType(res *recordedResponse) string {
	if res != nil && res.Headers != nil {
		if ct := normalizeContentType(res.Headers.Get("Content-Type")); ct != "" {
			return ct
		}
	}
	return applicationJSON
}

// normalizeContentType strips parameters (e.g. charset, boundary) from a content-type string.
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
