package gswag

import (
	"fmt"
	"os"
	"strings"

	openapi "github.com/swaggest/openapi-go"
)

// RegisterDSLOperation registers an operation declared via the rswag-style DSL.
// Called from a Ginkgo BeforeAll node so that spec registration happens once per
// operation, before any RunTest It blocks execute.
func (sc *SpecCollector) RegisterDSLOperation(op *dslOp) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if op.hidden || sc.isExcludedPath(op.path) {
		return
	}

	opCtx, err := sc.reflector.NewOperationContext(op.method, op.path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gswag DSL: NewOperationContext error for %s %s: %v\n", op.method, op.path, err)
		return
	}

	applyDSLOperationMeta(opCtx, op)

	// Path parameters from declared Parameter() calls.
	if pathStruct := buildPathParamsStructFromDSL(op.path, op.params); pathStruct != nil {
		opCtx.AddReqStructure(pathStruct)
	}

	if op.queryStruct != nil {
		opCtx.AddReqStructure(op.queryStruct)
	}

	if op.reqBodyModel != nil {
		if op.consumes != "" {
			opCtx.AddReqStructure(op.reqBodyModel, openapi.WithContentType(op.consumes))
		} else {
			opCtx.AddReqStructure(op.reqBodyModel)
		}
	}

	addDSLResponses(opCtx, op)

	for _, sec := range op.security {
		for name := range sec {
			sc.ensureSecurityScheme(name)
		}
	}

	if err := sc.reflector.AddOperation(opCtx); err != nil {
		fmt.Fprintf(os.Stderr, "gswag DSL: AddOperation error for %s %s: %v\n", op.method, op.path, err)
		return
	}

	sc.appendDSLParams(op)
	sc.appendDSLResponseHeaders(op)
}

// applyDSLOperationMeta copies metadata fields from the dslOp onto the operation context.
func applyDSLOperationMeta(opCtx openapi.OperationContext, op *dslOp) {
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
}

// addDSLResponses registers response structures on the operation context.
func addDSLResponses(opCtx openapi.OperationContext, op *dslOp) {
	if len(op.responses) == 0 {
		opCtx.AddRespStructure(nil, func(cu *openapi.ContentUnit) {
			cu.HTTPStatus = 200
		})
		return
	}

	for status, resp := range op.responses {
		s := status
		var model any
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
