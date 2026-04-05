package gswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/swaggest/openapi-go/openapi3"
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
