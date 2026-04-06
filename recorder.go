package gswag

import (
	"net/http"
	"time"
)

// recordedResponse captures the result of a do() call.
type recordedResponse struct {
	StatusCode int
	Headers    http.Header
	BodyBytes  []byte
	Duration   time.Duration

	builder          *requestBuilder // retained for spec registration
	RequestBodyBytes []byte          // request body bytes used for the request (if any)
}
