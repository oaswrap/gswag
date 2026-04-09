package output

import (
	"strings"

	"github.com/swaggest/openapi-go/openapi3"
)

// SanitizeSpecForSerialization normalizes operation parameters so spec marshalling
// is resilient to edge-cases that can leave Parameter.In empty.
func SanitizeSpecForSerialization(spec *openapi3.Spec) {
	if spec == nil || spec.Paths.MapOfPathItemValues == nil {
		return
	}

	for path, pathItem := range spec.Paths.MapOfPathItemValues {
		if pathItem.MapOfOperationValues == nil {
			continue
		}
		for method, op := range pathItem.MapOfOperationValues {
			if len(op.Parameters) == 0 {
				continue
			}
			op.Parameters = SanitizeOperationParameters(path, op.Parameters)
			pathItem.MapOfOperationValues[method] = op
		}
		spec.Paths.MapOfPathItemValues[path] = pathItem
	}
}

// SanitizeOperationParameters fixes missing `In` fields and deduplicates params.
func SanitizeOperationParameters(path string, params []openapi3.ParameterOrRef) []openapi3.ParameterOrRef {
	out := make([]openapi3.ParameterOrRef, 0, len(params))
	seen := make(map[string]struct{}, len(params))

	for _, por := range params {
		if por.Parameter == nil {
			out = append(out, por)
			continue
		}

		p := por.Parameter
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}

		in := strings.TrimSpace(string(p.In))
		if in == "" {
			if strings.Contains(path, "{"+name+"}") {
				p.In = openapi3.ParameterIn("path")
				req := true
				p.Required = &req
			} else {
				p.In = openapi3.ParameterIn("query")
			}
		}

		key := strings.ToLower(string(p.In) + "|" + name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		por.Parameter = p
		out = append(out, por)
	}

	return out
}
