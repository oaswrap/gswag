package output

import (
	"strings"

	"github.com/oaswrap/spec/openapi"
)

// SanitizeSpecForSerialization normalizes operation parameters before marshalling.
func SanitizeSpecForSerialization(spec *openapi.Document) {
	if spec == nil || spec.Paths == nil {
		return
	}
	for path, pathItem := range spec.Paths {
		if pathItem == nil {
			continue
		}
		sanitize := func(op *openapi.Operation) {
			if op != nil && len(op.Parameters) > 0 {
				op.Parameters = SanitizeOperationParameters(path, op.Parameters)
			}
		}
		sanitize(pathItem.Get)
		sanitize(pathItem.Put)
		sanitize(pathItem.Post)
		sanitize(pathItem.Delete)
		sanitize(pathItem.Options)
		sanitize(pathItem.Head)
		sanitize(pathItem.Patch)
		sanitize(pathItem.Trace)
		sanitize(pathItem.Query)
		for _, op := range pathItem.AdditionalOperations {
			sanitize(op)
		}
	}
}

// SanitizeOperationParameters fixes missing `in` fields and deduplicates params.
func SanitizeOperationParameters(path string, params []*openapi.Parameter) []*openapi.Parameter {
	out := make([]*openapi.Parameter, 0, len(params))
	seen := make(map[string]struct{}, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		in := strings.TrimSpace(p.In)
		if in == "" {
			if strings.Contains(path, "{"+name+"}") {
				p.In = string(openapi.ParameterInPath)
				p.Required = true
			} else {
				p.In = string(openapi.ParameterInQuery)
			}
		}
		key := strings.ToLower(p.In + "|" + name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}
