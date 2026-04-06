package gswag

import (
	"net/http"
	"time"
)

// RecordedResponse captures the result of a Do() call and exposes
// fields consumed by Gomega matchers.
type RecordedResponse struct {
	StatusCode int
	Headers    http.Header
	BodyBytes  []byte
	Duration   time.Duration

	// retained so spec registration can access metadata
	builder *RequestBuilder
	// RequestBodyBytes stores the request body bytes used for the request (if any)
	RequestBodyBytes []byte
}
