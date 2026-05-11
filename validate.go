package gswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/oaswrap/spec/openapi"
	"github.com/xeipuuv/gojsonschema"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// ValidationIssue describes a single spec problem.
type ValidationIssue struct {
	Severity string // SeverityError or SeverityWarning
	Path     string // e.g. "paths./users.get"
	Message  string
}

func (v ValidationIssue) String() string {
	return fmt.Sprintf("[%s] %s: %s", strings.ToUpper(v.Severity), v.Path, v.Message)
}

// ValidateSpec runs structural validation on the collected spec and returns any issues found.
func ValidateSpec() []ValidationIssue {
	if globalCollector == nil {
		return []ValidationIssue{
			{Severity: SeverityError, Path: "", Message: "gswag not initialised — call Init() first"},
		}
	}
	globalCollector.mu.Lock()
	specDoc := globalCollector.doc
	globalCollector.mu.Unlock()
	return validateSpec(specDoc)
}

// ValidateSpecFile reads a YAML or JSON spec file and runs structural validation.
func ValidateSpecFile(path string) ([]ValidationIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	specDoc := &openapi.Document{}
	if err := yaml.Unmarshal(data, specDoc); err != nil {
		if err2 := json.Unmarshal(data, specDoc); err2 != nil {
			return nil, fmt.Errorf("parsing spec: yaml: %w; json: %w", err, err2)
		}
	}
	return validateSpec(specDoc), nil
}

func validateSpec(specDoc *openapi.Document) []ValidationIssue {
	var issues []ValidationIssue
	add := func(severity, path, msg string) {
		issues = append(issues, ValidationIssue{Severity: severity, Path: path, Message: msg})
	}
	if specDoc == nil {
		add(SeverityError, "", "spec is nil")
		return issues
	}
	if specDoc.Info.Title == "" {
		add(SeverityError, "info.title", "title is required")
	}
	if specDoc.Info.Version == "" {
		add(SeverityError, "info.version", "version is required")
	}
	if len(specDoc.Paths) == 0 {
		add(SeverityWarning, "paths", "no paths defined")
	}
	declaredSchemes := map[string]bool{}
	if specDoc.Components != nil {
		for name := range specDoc.Components.SecuritySchemes {
			declaredSchemes[name] = true
		}
	}
	forEachPathOperation(specDoc, func(path, method string, op *openapi.Operation) {
		loc := fmt.Sprintf("paths.%s.%s", path, method)
		if op.Summary == "" {
			add(SeverityWarning, loc, "operation has no summary")
		}
		if len(op.Tags) == 0 {
			add(SeverityWarning, loc, "operation has no tags")
		}
		for _, secReq := range op.Security {
			for name := range secReq {
				if !declaredSchemes[name] {
					add(
						SeverityError,
						loc,
						fmt.Sprintf("security scheme %q is not declared in components/securitySchemes", name),
					)
				}
			}
		}
	})
	return issues
}

// ErrSpecInvalid is returned when the spec has at least one error-level issue.
var ErrSpecInvalid = errors.New("spec has validation errors")

// WriteAndValidateSpec writes the spec and then validates it.
func WriteAndValidateSpec() error {
	if err := WriteSpec(); err != nil {
		return err
	}
	issues := ValidateSpec()
	var errs []string
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			errs = append(errs, issue.String())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%w:\n%s", ErrSpecInvalid, strings.Join(errs, "\n"))
	}
	return nil
}

// validateResponseAgainstOperation validates an actual recordedResponse against the declared response model/schema.
func validateResponseAgainstOperation(b *requestBuilder, res *recordedResponse) ([]string, error) {
	if b == nil {
		return nil, errors.New("nil requestBuilder")
	}
	if model, ok := b.respBodies[res.StatusCode]; ok && model != nil {
		mt := reflect.TypeOf(model)
		var target reflect.Value
		if mt.Kind() == reflect.Ptr {
			target = reflect.New(mt.Elem())
		} else {
			target = reflect.New(mt)
		}
		if err := json.Unmarshal(res.BodyBytes, target.Interface()); err != nil {
			return []string{fmt.Sprintf("unmarshal typed model: %v", err)}, nil
		}
		return nil, nil
	}
	if globalCollector == nil {
		return nil, errors.New("spec not initialised")
	}
	globalCollector.mu.Lock()
	specDoc := globalCollector.doc
	globalCollector.mu.Unlock()

	pathItem := specDoc.Paths[b.path]
	if pathItem == nil {
		return nil, nil
	}
	op, ok := pathItemOperation(pathItem, b.method)
	if !ok {
		return nil, nil
	}
	statusKey := strconv.Itoa(res.StatusCode)
	resp := op.Responses[statusKey]
	if resp == nil {
		resp = op.Responses["default"]
	}
	if resp == nil {
		return nil, nil
	}
	var media *openapi.MediaType
	if resp.Content != nil {
		if m, found := resp.Content[applicationJSON]; found {
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
	schemaBytes, err := json.Marshal(schemaDocument(media.Schema, specDoc))
	if err != nil {
		return nil, fmt.Errorf("marshal json schema: %w", err)
	}
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	docLoader := gojsonschema.NewBytesLoader(res.BodyBytes)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return nil, err
	}
	if result.Valid() {
		return nil, nil
	}
	issues := make([]string, 0, len(result.Errors()))
	for _, e := range result.Errors() {
		issues = append(issues, e.String())
	}
	return issues, nil
}

func schemaDocument(schema *openapi.Schema, specDoc *openapi.Document) map[string]any {
	raw, _ := openapi.MarshalJSON(schema)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		out = map[string]any{}
	}
	if specDoc != nil && specDoc.Components != nil && len(specDoc.Components.Schemas) > 0 {
		rawComponents, _ := openapi.MarshalJSON(specDoc.Components)
		var components any
		if json.Unmarshal(rawComponents, &components) == nil {
			out["components"] = components
		}
	}
	return out
}
