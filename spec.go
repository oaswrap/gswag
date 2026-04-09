package gswag

import (
	"regexp"
	"sync"

	"github.com/swaggest/openapi-go/openapi3"
)

// bearerAuthSchemeName is the conventional component key for Bearer JWT schemes.
const bearerAuthSchemeName = "bearerAuth"

// applicationJSON is the default JSON media type used across the package.
const applicationJSON = "application/json"

// pathParamRe matches OpenAPI path parameters like {id}.
var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

// genericInstRe detects Go generic instantiation names like "Page[pkg/path.Item]".
var genericInstRe = regexp.MustCompile(`^(\w+)\[(.+)\]$`)

// SpecCollector accumulates OpenAPI operations from test executions in a thread-safe manner.
type SpecCollector struct {
	mu           sync.Mutex
	reflector    *openapi3.Reflector
	excludePaths []string
}
