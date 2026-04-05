package gswag

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// RequestBuilder constructs an HTTP request and its OpenAPI operation metadata
// using a fluent DSL.
type RequestBuilder struct {
	method          string
	path            string
	pathParams      map[string]string
	queryParams     map[string]string
	headers         map[string]string
	body            interface{}   // typed struct → schema inference
	bodyRaw         []byte        // raw JSON body fallback
	bodyContentType string        // for raw body
	respBodies      map[int]interface{}
	queryStruct     interface{}   // typed struct with `query` tags → query param schemas
	tags            []string
	summary         string
	description     string
	operationID     string
	security        []map[string][]string
	deprecated      bool
}

// GET creates a RequestBuilder for a GET request.
func GET(path string) *RequestBuilder { return newBuilder(http.MethodGet, path) }

// POST creates a RequestBuilder for a POST request.
func POST(path string) *RequestBuilder { return newBuilder(http.MethodPost, path) }

// PUT creates a RequestBuilder for a PUT request.
func PUT(path string) *RequestBuilder { return newBuilder(http.MethodPut, path) }

// PATCH creates a RequestBuilder for a PATCH request.
func PATCH(path string) *RequestBuilder { return newBuilder(http.MethodPatch, path) }

// DELETE creates a RequestBuilder for a DELETE request.
func DELETE(path string) *RequestBuilder { return newBuilder(http.MethodDelete, path) }

func newBuilder(method, path string) *RequestBuilder {
	return &RequestBuilder{
		method:      method,
		path:        path,
		pathParams:  make(map[string]string),
		queryParams: make(map[string]string),
		headers:     make(map[string]string),
		respBodies:  make(map[int]interface{}),
	}
}

// WithSummary sets the operation summary.
func (b *RequestBuilder) WithSummary(s string) *RequestBuilder {
	b.summary = s
	return b
}

// WithDescription sets the operation description.
func (b *RequestBuilder) WithDescription(s string) *RequestBuilder {
	b.description = s
	return b
}

// WithTag appends tags to the operation.
func (b *RequestBuilder) WithTag(tags ...string) *RequestBuilder {
	b.tags = append(b.tags, tags...)
	return b
}

// WithOperationID sets the operationId.
func (b *RequestBuilder) WithOperationID(id string) *RequestBuilder {
	b.operationID = id
	return b
}

// WithPathParam registers a path parameter. The path template must contain {key}.
func (b *RequestBuilder) WithPathParam(key, value string) *RequestBuilder {
	b.pathParams[key] = value
	return b
}

// WithQueryParam adds a query parameter.
func (b *RequestBuilder) WithQueryParam(key, value string) *RequestBuilder {
	b.queryParams[key] = value
	return b
}

// WithHeader adds a request header.
func (b *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	b.headers[key] = value
	return b
}

// WithQueryParamStruct registers a typed struct whose fields carry `query` struct tags.
// These fields are reflected into OpenAPI query parameter definitions.
//
// Example:
//
//	type ListQuery struct {
//	    Page  int    `query:"page"`
//	    Limit int    `query:"limit"`
//	    Sort  string `query:"sort"`
//	}
//	gswag.GET("/items").WithQueryParamStruct(new(ListQuery)).Do(srv)
func (b *RequestBuilder) WithQueryParamStruct(v interface{}) *RequestBuilder {
	b.queryStruct = v
	return b
}

// WithRequestBody sets a typed struct as the request body.
// The struct is used for schema inference and as the actual JSON body.
func (b *RequestBuilder) WithRequestBody(body interface{}) *RequestBuilder {
	b.body = body
	return b
}

// WithRawBody sets a raw byte slice as the request body.
func (b *RequestBuilder) WithRawBody(body []byte, contentType string) *RequestBuilder {
	b.bodyRaw = body
	b.bodyContentType = contentType
	return b
}

// ExpectResponseBody registers the typed struct for the default (200) response schema.
func (b *RequestBuilder) ExpectResponseBody(model interface{}) *RequestBuilder {
	b.respBodies[http.StatusOK] = model
	return b
}

// ExpectResponseBodyFor registers a typed struct for a specific HTTP status response schema.
func (b *RequestBuilder) ExpectResponseBodyFor(status int, model interface{}) *RequestBuilder {
	b.respBodies[status] = model
	return b
}

// WithBearerAuth adds Bearer JWT authentication to the operation.
// A "bearerAuth" HTTP Bearer scheme with BearerFormat: JWT is auto-registered
// in the spec components if not already present.
func (b *RequestBuilder) WithBearerAuth() *RequestBuilder {
	b.security = append(b.security, map[string][]string{"bearerAuth": {}})
	return b
}

// WithSecurity adds a named security requirement to the operation.
func (b *RequestBuilder) WithSecurity(schemeName string, scopes ...string) *RequestBuilder {
	b.security = append(b.security, map[string][]string{schemeName: scopes})
	return b
}

// AsDeprecated marks the operation as deprecated.
func (b *RequestBuilder) AsDeprecated() *RequestBuilder {
	b.deprecated = true
	return b
}

// Do executes the HTTP request against target (either *httptest.Server or a base URL string),
// records the response, and registers the operation with the global SpecCollector.
func (b *RequestBuilder) Do(target interface{}) *RecordedResponse {
	baseURL := resolveBaseURL(target)
	url := baseURL + b.resolvedPath()

	req, err := b.buildRequest(url)
	if err != nil {
		panic("gswag: failed to build request: " + err.Error())
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		panic("gswag: HTTP request failed: " + err.Error())
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("gswag: failed to read response body: " + err.Error())
	}

	recorded := &RecordedResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		BodyBytes:  body,
		Duration:   duration,
		builder:    b,
	}

	if globalCollector != nil {
		globalCollector.Register(b, recorded)
	}

	return recorded
}

// resolvedPath replaces path param templates with their concrete values.
func (b *RequestBuilder) resolvedPath() string {
	p := b.path
	for k, v := range b.pathParams {
		p = strings.ReplaceAll(p, "{"+k+"}", v)
	}
	return p
}

func (b *RequestBuilder) buildRequest(url string) (*http.Request, error) {
	var bodyReader io.Reader

	contentType := "application/json"
	if b.body != nil {
		data, err := json.Marshal(b.body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	} else if len(b.bodyRaw) > 0 {
		bodyReader = bytes.NewReader(b.bodyRaw)
		if b.bodyContentType != "" {
			contentType = b.bodyContentType
		}
	}

	req, err := http.NewRequest(b.method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	if len(b.queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range b.queryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func resolveBaseURL(target interface{}) string {
	switch t := target.(type) {
	case *httptest.Server:
		return t.URL
	case string:
		return strings.TrimRight(t, "/")
	default:
		panic("gswag: Do() expects *httptest.Server or string as target")
	}
}
