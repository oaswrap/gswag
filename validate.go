package gswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/swaggest/openapi-go/openapi3"
	"github.com/xeipuuv/gojsonschema"
)

// ValidationIssue describes a single spec problem.
type ValidationIssue struct {
	Severity string // "error" or "warning"
	Path     string // e.g. "paths./users.get"
	Message  string
}

func (v ValidationIssue) String() string {
	return fmt.Sprintf("[%s] %s: %s", strings.ToUpper(v.Severity), v.Path, v.Message)
}

// ValidateSpec runs structural validation on the collected spec and returns any issues found.
// Errors must be fixed for a valid spec; warnings are informational.
func ValidateSpec() []ValidationIssue {
	if globalCollector == nil {
		return []ValidationIssue{{Severity: "error", Path: "", Message: "gswag not initialised — call Init() first"}}
	}
	globalCollector.mu.Lock()
	spec := globalCollector.reflector.Spec
	globalCollector.mu.Unlock()
	return validateSpec(spec)
}

// ValidateSpecFile reads a YAML or JSON spec file and runs structural validation.
func ValidateSpecFile(path string) ([]ValidationIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	spec := &openapi3.Spec{}
	if err := spec.UnmarshalYAML(data); err != nil {
		// Try JSON fallback.
		if err2 := json.Unmarshal(data, spec); err2 != nil {
			return nil, fmt.Errorf("parsing spec: yaml: %v; json: %v", err, err2)
		}
	}
	return validateSpec(spec), nil
}

func validateSpec(spec *openapi3.Spec) []ValidationIssue {
	var issues []ValidationIssue

	add := func(severity, path, msg string) {
		issues = append(issues, ValidationIssue{Severity: severity, Path: path, Message: msg})
	}

	// Info checks.
	if spec.Info.Title == "" {
		add("error", "info.title", "title is required")
	}
	if spec.Info.Version == "" {
		add("error", "info.version", "version is required")
	}

	// Paths checks.
	if len(spec.Paths.MapOfPathItemValues) == 0 {
		add("warning", "paths", "no paths defined")
	}

	// Collect declared security scheme names.
	declaredSchemes := map[string]bool{}
	if spec.Components != nil && spec.Components.SecuritySchemes != nil {
		for name := range spec.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues {
			declaredSchemes[name] = true
		}
	}

	for path, item := range spec.Paths.MapOfPathItemValues {
		for method, op := range item.MapOfOperationValues {
			loc := fmt.Sprintf("paths.%s.%s", path, method)

			if op.Summary == nil || *op.Summary == "" {
				add("warning", loc, "operation has no summary")
			}
			if len(op.Tags) == 0 {
				add("warning", loc, "operation has no tags")
			}

			// Check all security requirements reference declared schemes.
			for _, secReq := range op.Security {
				for name := range secReq {
					if !declaredSchemes[name] {
						add("error", loc, fmt.Sprintf("security scheme %q is not declared in components/securitySchemes", name))
					}
				}
			}
		}
	}

	return issues
}

// ErrSpecInvalid is returned when the spec has at least one error-level issue.
var ErrSpecInvalid = errors.New("spec has validation errors")

// WriteAndValidateSpec writes the spec and then validates it.
// Returns ErrSpecInvalid (wrapping the issue list) if any errors are found.
func WriteAndValidateSpec() error {
	if err := WriteSpec(); err != nil {
		return err
	}
	issues := ValidateSpec()
	var errs []string
	for _, issue := range issues {
		if issue.Severity == "error" {
			errs = append(errs, issue.String())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%w:\n%s", ErrSpecInvalid, strings.Join(errs, "\n"))
	}
	return nil
}

// validateResponseAgainstOperation validates an actual recordedResponse against
// the declared typed response model on the requestBuilder (if present). It
// attempts typed-model unmarshalling first, and falls back to full JSON Schema
// validation generated from the collected spec using gojsonschema.
func validateResponseAgainstOperation(b *requestBuilder, res *recordedResponse) ([]string, error) {
	if b == nil {
		return nil, fmt.Errorf("nil requestBuilder")
	}

	// If the builder declared a typed response model for this status code,
	// validate by attempting to unmarshal the response JSON into that type.
	if model, ok := b.respBodies[res.StatusCode]; ok && model != nil {
		mt := reflect.TypeOf(model)
		var target reflect.Value
		if mt.Kind() == reflect.Ptr {
			target = reflect.New(mt.Elem())
		} else {
			target = reflect.New(mt)
		}
		if err := json.Unmarshal(res.BodyBytes, target.Interface()); err != nil {
			return []string{err.Error()}, nil
		}
		return nil, nil
	}

	// No declared typed model — attempt full JSON Schema validation
	if globalCollector == nil || globalCollector.reflector == nil {
		return nil, fmt.Errorf("spec not initialised")
	}

	globalCollector.mu.Lock()
	spec := globalCollector.reflector.Spec
	globalCollector.mu.Unlock()

	// Locate operation in the collected spec using builder's method + path template.
	pathItem, ok := spec.Paths.MapOfPathItemValues[b.path]
	if !ok {
		// Path not found in spec — nothing to validate against.
		return nil, nil
	}

	op, ok := pathItem.MapOfOperationValues[strings.ToLower(b.method)]
	if !ok {
		return nil, nil
	}

	statusKey := fmt.Sprintf("%d", res.StatusCode)

	var resp *openapi3.Response
	if op.Responses.MapOfResponseOrRefValues != nil {
		if ror, found := op.Responses.MapOfResponseOrRefValues[statusKey]; found {
			resp = ror.Response
		}
	}
	if resp == nil && op.Responses.Default != nil {
		resp = op.Responses.Default.Response
	}
	if resp == nil {
		return nil, nil
	}

	// Prefer application/json content, otherwise take the first available content.
	var media *openapi3.MediaType
	if resp.Content != nil {
		if m, found := resp.Content["application/json"]; found {
			media = &m
		} else {
			for _, m := range resp.Content {
				media = &m
				break
			}
		}
	}
	if media == nil || media.Schema == nil {
		return nil, nil
	}

	// Convert OpenAPI SchemaOrRef to JSON Schema structure (inlines components as needed).
	js := media.Schema.ToJSONSchema(spec)

	schemaBytes, err := json.Marshal(js)
	if err != nil {
		return nil, fmt.Errorf("marshal json schema: %w", err)
	}

	// Compile schema using gojsonschema and validate the instance bytes.
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	schemaCompiled, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(res.BodyBytes)
	result, err := schemaCompiled.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failure: %w", err)
	}

	if !result.Valid() {
		var msgs []string
		for _, e := range result.Errors() {
			msgs = append(msgs, e.String())
		}
		return msgs, nil
	}

	return nil, nil
}
