package gswag

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
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
	doc          *openapi.Document
	openapiOpts  []option.OpenAPIOption
	excludePaths []string
}

func (sc *SpecCollector) operation(method, path string) (*openapi.Operation, bool) {
	if sc == nil || sc.doc == nil || sc.doc.Paths == nil {
		return nil, false
	}
	item := sc.doc.Paths[path]
	if item == nil {
		return nil, false
	}
	return pathItemOperation(item, method)
}

func pathItemOperation(item *openapi.PathItem, method string) (*openapi.Operation, bool) {
	if item == nil {
		return nil, false
	}
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return item.Get, item.Get != nil
	case http.MethodPut:
		return item.Put, item.Put != nil
	case http.MethodPost:
		return item.Post, item.Post != nil
	case http.MethodDelete:
		return item.Delete, item.Delete != nil
	case http.MethodOptions:
		return item.Options, item.Options != nil
	case http.MethodHead:
		return item.Head, item.Head != nil
	case http.MethodPatch:
		return item.Patch, item.Patch != nil
	case http.MethodTrace:
		return item.Trace, item.Trace != nil
	case "QUERY":
		return item.Query, item.Query != nil
	default:
		if item.AdditionalOperations == nil {
			return nil, false
		}
		op := item.AdditionalOperations[strings.ToUpper(method)]
		return op, op != nil
	}
}

func forEachPathOperation(doc *openapi.Document, fn func(path, method string, op *openapi.Operation)) {
	if doc == nil || doc.Paths == nil {
		return
	}
	for path, item := range doc.Paths {
		if item == nil {
			continue
		}
		if item.Get != nil {
			fn(path, strings.ToLower(http.MethodGet), item.Get)
		}
		if item.Put != nil {
			fn(path, strings.ToLower(http.MethodPut), item.Put)
		}
		if item.Post != nil {
			fn(path, strings.ToLower(http.MethodPost), item.Post)
		}
		if item.Delete != nil {
			fn(path, strings.ToLower(http.MethodDelete), item.Delete)
		}
		if item.Options != nil {
			fn(path, strings.ToLower(http.MethodOptions), item.Options)
		}
		if item.Head != nil {
			fn(path, strings.ToLower(http.MethodHead), item.Head)
		}
		if item.Patch != nil {
			fn(path, strings.ToLower(http.MethodPatch), item.Patch)
		}
		if item.Trace != nil {
			fn(path, strings.ToLower(http.MethodTrace), item.Trace)
		}
		if item.Query != nil {
			fn(path, "query", item.Query)
		}
		for method, op := range item.AdditionalOperations {
			fn(path, strings.ToLower(method), op)
		}
	}
}

func registerTempOperation(
	opts []option.OpenAPIOption,
	method, path string,
	opOpts ...option.OperationOption,
) *openapi.Document {
	r := spec.NewRouter(opts...)
	r.Add(method, path, opOpts...)
	return r.Document()
}
