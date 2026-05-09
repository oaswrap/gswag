package gswag

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/oaswrap/gswag/internal/schemautil"
	"github.com/oaswrap/spec/openapi"
)

// injectInferredRequestSchema attaches an inferred request-body schema from actual request bytes.
func (sc *SpecCollector) injectInferredRequestSchema(b *requestBuilder, res *recordedResponse) {
	if len(res.RequestBodyBytes) == 0 {
		return
	}
	op, ok := sc.operation(b.method, b.path)
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
	var schema *openapi.Schema
	if strings.Contains(strings.ToLower(ct), "json") {
		schema = schemautil.InferSchema(res.RequestBodyBytes)
	}
	if schema == nil {
		schema = &openapi.Schema{Type: string(String), Format: "binary"}
	}
	if op.RequestBody == nil {
		op.RequestBody = &openapi.RequestBody{Content: map[string]openapi.MediaType{}}
	}
	if op.RequestBody.Content == nil {
		op.RequestBody.Content = map[string]openapi.MediaType{}
	}
	for _, existingMT := range op.RequestBody.Content {
		if existingMT.Schema != nil {
			return
		}
	}
	mt := op.RequestBody.Content[ct]
	if mt.Schema == nil {
		mt.Schema = schema
		op.RequestBody.Content[ct] = mt
	}
}

// injectInferredSchema attaches a best-effort response schema to an existing response slot.
func (sc *SpecCollector) injectInferredSchema(b *requestBuilder, res *recordedResponse) {
	inferred := schemautil.InferSchema(res.BodyBytes)
	if inferred == nil {
		return
	}
	op, ok := sc.operation(b.method, b.path)
	if !ok {
		return
	}
	statusKey := strconv.Itoa(res.StatusCode)
	resp := op.Responses[statusKey]
	if resp == nil {
		return
	}
	ct := applicationJSON
	if resp.Content == nil {
		resp.Content = map[string]openapi.MediaType{ct: {Schema: inferred}}
		return
	}
	mt := resp.Content[ct]
	if mt.Schema == nil {
		mt.Schema = inferred
		resp.Content[ct] = mt
	}
}

// injectRecordedResponseSchema injects an inferred schema from the actual response body.
func (sc *SpecCollector) injectRecordedResponseSchema(method, path string, res *recordedResponse) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	op, ok := sc.operation(method, path)
	if !ok {
		return
	}
	statusKey := strconv.Itoa(res.StatusCode)
	resp := op.Responses[statusKey]
	if resp == nil {
		return
	}
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
		resp.Content = map[string]openapi.MediaType{ct: {Schema: inferred}}
	} else {
		mt := resp.Content[ct]
		if mt.Schema == nil {
			mt.Schema = inferred
			resp.Content[ct] = mt
		}
	}
}

// appendResponseHeaders attaches declared response header schemas from requestBuilder.
func (sc *SpecCollector) appendResponseHeaders(b *requestBuilder) {
	if len(b.respHeaders) == 0 {
		return
	}
	op, ok := sc.operation(b.method, b.path)
	if !ok {
		return
	}
	for status, headers := range b.respHeaders {
		resp := op.Responses[strconv.Itoa(status)]
		if resp == nil {
			continue
		}
		if resp.Headers == nil {
			resp.Headers = make(map[string]*openapi.Header)
		}
		for name, model := range headers {
			resp.Headers[name] = buildHeaderOrRef(model)
		}
	}
}

// appendDSLResponseHeaders attaches response header schemas declared via ResponseHeader().
func (sc *SpecCollector) appendDSLResponseHeaders(op *dslOp) {
	operation, ok := sc.operation(op.method, op.path)
	if !ok {
		return
	}
	for status, resp := range op.responses {
		if resp == nil || len(resp.headers) == 0 {
			continue
		}
		r := operation.Responses[strconv.Itoa(status)]
		if r == nil {
			continue
		}
		if r.Headers == nil {
			r.Headers = make(map[string]*openapi.Header)
		}
		for name, model := range resp.headers {
			r.Headers[name] = buildHeaderOrRef(model)
		}
	}
}

func buildHeaderOrRef(model any) *openapi.Header {
	return &openapi.Header{Schema: inferHeaderSchema(model)}
}

func inferHeaderSchema(model any) *openapi.Schema {
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

func stringSchemaOrRef() *openapi.Schema {
	return &openapi.Schema{Type: string(String)}
}

// appendExamples attaches captured request/response examples to the spec.
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
	op, ok := sc.operation(b.method, b.path)
	if !ok {
		return
	}
	if res.StatusCode/100 == 2 && op.RequestBody != nil && len(res.RequestBodyBytes) > 0 {
		ct := requestExampleContentType(b, res)
		if op.RequestBody.Content != nil {
			if mt, found := op.RequestBody.Content[ct]; found {
				bts := sanitize(res.RequestBodyBytes)
				var ex any
				if err := json.Unmarshal(bts, &ex); err == nil {
					mt.Example = ex
					op.RequestBody.Content[ct] = mt
				}
			}
		}
	}
	statusKey := strconv.Itoa(res.StatusCode)
	if resp := op.Responses[statusKey]; resp != nil {
		ct := responseExampleContentType(res)
		if resp.Content != nil {
			if mt, found := resp.Content[ct]; found {
				bts := sanitize(res.BodyBytes)
				var ex any
				if err := json.Unmarshal(bts, &ex); err == nil {
					mt.Example = ex
					resp.Content[ct] = mt
				}
			}
		}
	}
}

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

func responseExampleContentType(res *recordedResponse) string {
	if res != nil && res.Headers != nil {
		if ct := normalizeContentType(res.Headers.Get("Content-Type")); ct != "" {
			return ct
		}
	}
	return applicationJSON
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
