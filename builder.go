package gswag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

// requestBuilder constructs an HTTP request and its OpenAPI operation metadata.
// It is an internal type used by the DSL; users interact through the DSL functions.
type requestBuilder struct {
	method          string
	path            string
	pathParams      map[string]string
	queryParams     map[string]string
	headers         map[string]string
	body            any    // typed struct → schema inference
	bodyRaw         []byte // raw JSON body fallback
	bodyContentType string // for raw body
	respBodies      map[int]any
	respHeaders     map[int]map[string]any
	queryStruct     any // typed struct with `query` tags → query param schemas
	tags            []string
	summary         string
	description     string
	operationID     string
	security        []map[string][]string
	deprecated      bool
}

func newRequestBuilder(method, path string) *requestBuilder {
	return &requestBuilder{
		method:      method,
		path:        path,
		pathParams:  make(map[string]string),
		queryParams: make(map[string]string),
		headers:     make(map[string]string),
		respBodies:  make(map[int]any),
		respHeaders: make(map[int]map[string]any),
	}
}

// do executes the HTTP request against target (*httptest.Server or base URL string)
// and records the response. The caller decides whether to register the result with
// the spec collector.
func (b *requestBuilder) do(target any) *recordedResponse {
	baseURL := resolveBaseURL(target)
	url := baseURL + b.resolvedPath()

	req, reqBodyBytes, err := b.buildRequest(url)
	if err != nil {
		panic("gswag: failed to build request: " + err.Error())
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	duration := time.Since(start)
	if err != nil {
		panic("gswag: HTTP request failed: " + err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("gswag: failed to read response body: " + err.Error())
	}

	recorded := &recordedResponse{
		StatusCode:       resp.StatusCode,
		Headers:          resp.Header,
		BodyBytes:        body,
		Duration:         duration,
		builder:          b,
		RequestBodyBytes: reqBodyBytes,
	}

	if globalCollector != nil && (b.summary != "" || len(b.tags) > 0 || len(b.respBodies) > 0) {
		if globalConfig != nil && globalConfig.EnforceResponseValidation {
			issues, verr := validateResponseAgainstOperation(b, recorded)
			warnMode := strings.EqualFold(globalConfig.ValidationMode, "warn")
			if verr != nil {
				if warnMode {
					fmt.Fprintln(os.Stderr, "gswag: response validation error:", verr)
				} else {
					panic("gswag: response validation error: " + verr.Error())
				}
			} else if len(issues) > 0 {
				msg := "gswag: response does not match declared schema: " + strings.Join(issues, "; ")
				if warnMode {
					fmt.Fprintln(os.Stderr, msg)
				} else {
					panic(msg)
				}
			}
		}
		globalCollector.Register(b, recorded)
	}

	return recorded
}

// resolvedPath replaces path param templates with their concrete values.
func (b *requestBuilder) resolvedPath() string {
	p := b.path
	for k, v := range b.pathParams {
		p = strings.ReplaceAll(p, "{"+k+"}", v)
	}
	return p
}

func (b *requestBuilder) buildRequest(url string) (*http.Request, []byte, error) {
	var bodyReader io.Reader
	var data []byte

	contentType := applicationJSON
	if b.body != nil {
		d, err := json.Marshal(b.body)
		if err != nil {
			return nil, nil, err
		}
		data = d
		bodyReader = bytes.NewReader(d)
	} else if len(b.bodyRaw) > 0 {
		data = b.bodyRaw
		bodyReader = bytes.NewReader(b.bodyRaw)
		if b.bodyContentType != "" {
			contentType = b.bodyContentType
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), b.method, url, bodyReader)
	if err != nil {
		return nil, nil, err
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

	return req, data, nil
}

func resolveBaseURL(target any) string {
	switch t := target.(type) {
	case *httptest.Server:
		return t.URL
	case string:
		return strings.TrimRight(t, "/")
	default:
		panic("gswag: do() expects *httptest.Server or string as target")
	}
}
