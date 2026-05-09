package gswag

import (
	"net/http"
	"strings"

	"github.com/oaswrap/spec/option"
)

// RegisterDSLOperation registers an operation declared via the rswag-style DSL.
func (sc *SpecCollector) RegisterDSLOperation(op *dslOp) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if op.hidden || sc.isExcludedPath(op.path) {
		return
	}

	doc := registerTempOperation(sc.openapiOpts, op.method, op.path, buildDSLOperationOptions(op)...)
	mergeSpec(sc.doc, doc)

	for _, sec := range op.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	sc.appendDSLParams(op)
	sc.appendDSLResponseHeaders(op)
}

func buildDSLOperationOptions(op *dslOp) []option.OperationOption {
	var opts []option.OperationOption
	if len(op.tags) > 0 {
		opts = append(opts, option.Tags(op.tags...))
	}
	if op.summary != "" {
		opts = append(opts, option.Summary(op.summary))
	}
	if op.description != "" {
		opts = append(opts, option.Description(op.description))
	}
	if op.operationID != "" {
		opts = append(opts, option.OperationID(op.operationID))
	}
	if op.deprecated {
		opts = append(opts, option.Deprecated())
	}
	for _, sec := range op.security {
		for name, scopes := range sec {
			opts = append(opts, option.Security(name, scopes...))
		}
	}
	if pathStruct := buildPathParamsStructFromDSL(op.path, op.params); pathStruct != nil {
		opts = append(opts, option.Request(pathStruct))
	}
	if op.queryStruct != nil {
		opts = append(opts, option.Request(op.queryStruct))
	}
	if op.reqBodyModel != nil {
		if op.consumes != "" {
			opts = append(opts, option.Request(op.reqBodyModel, option.ContentType(op.consumes)))
		} else {
			opts = append(opts, option.Request(op.reqBodyModel))
		}
	}
	if len(op.responses) == 0 {
		opts = append(opts, option.Response(http.StatusOK, nil))
		return opts
	}
	for status, resp := range op.responses {
		var model any
		var contentOpts []option.ContentOption
		if resp != nil {
			model = resp.bodyModel
			if resp.description != "" {
				contentOpts = append(contentOpts, option.ContentDescription(resp.description))
			}
		}
		if len(op.produces) > 0 {
			for _, ct := range op.produces {
				perContentOpts := append([]option.ContentOption{}, contentOpts...)
				perContentOpts = append(perContentOpts, option.ContentType(ct))
				opts = append(opts, option.Response(status, model, perContentOpts...))
			}
		} else {
			opts = append(opts, option.Response(status, model, contentOpts...))
		}
	}
	return opts
}

// isExcludedPath reports whether path matches any pattern in sc.excludePaths.
func (sc *SpecCollector) isExcludedPath(path string) bool {
	for _, pattern := range sc.excludePaths {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if pattern == path {
			return true
		}
		if before, ok := strings.CutSuffix(pattern, "*"); ok {
			if strings.HasPrefix(path, before) {
				return true
			}
		}
	}
	return false
}
