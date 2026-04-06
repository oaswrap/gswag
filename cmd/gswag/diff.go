package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/swaggest/openapi-go/openapi3"
)

type diffResult struct {
	added    []string
	removed  []string
	modified []string
}

func runDiff(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gswag diff [--json] [--no-fail] <base-spec> <new-spec>")
		os.Exit(1)
	}

	jsonOut := hasFlag(args, "--json") || hasFlag(args, "--format=json")
	failOnBreaking := !hasFlag(args, "--no-fail")

	base, err := loadSpec(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading base spec: %v\n", err)
		os.Exit(2)
	}
	next, err := loadSpec(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading new spec: %v\n", err)
		os.Exit(2)
	}

	result := diffSpecs(base, next)
	if jsonOut {
		out := struct {
			Added    []string `json:"added"`
			Removed  []string `json:"removed"`
			Modified []string `json:"modified"`
			Breaking bool     `json:"breaking"`
		}{
			Added:    result.added,
			Removed:  result.removed,
			Modified: result.modified,
			Breaking: containsBreakingResult(result),
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		if out.Breaking && failOnBreaking {
			os.Exit(1)
		}
		return
	}

	hasBreaking := printDiff(result)
	if hasBreaking && failOnBreaking {
		os.Exit(1)
	}
}

func containsBreakingResult(r diffResult) bool {
	if len(r.removed) > 0 {
		return true
	}
	for _, m := range r.modified {
		if containsBreaking(m) {
			return true
		}
	}
	return false
}

func loadSpec(path string) (*openapi3.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	spec := &openapi3.Spec{}
	// Try YAML first, then JSON.
	if err := spec.UnmarshalYAML(data); err != nil {
		if err2 := json.Unmarshal(data, spec); err2 != nil {
			return nil, fmt.Errorf("yaml: %v; json: %v", err, err2)
		}
	}
	return spec, nil
}

func diffSpecs(base, next *openapi3.Spec) diffResult {
	var r diffResult

	basePaths := pathMethodSet(base)
	nextPaths := pathMethodSet(next)

	// Removed paths/methods → breaking.
	for key := range basePaths {
		if !nextPaths[key] {
			r.removed = append(r.removed, key)
		}
	}

	// Added paths/methods → non-breaking.
	for key := range nextPaths {
		if !basePaths[key] {
			r.added = append(r.added, key)
		}
	}

	// Modified — compare parameters and response status codes for shared operations.
	for path, baseItem := range base.Paths.MapOfPathItemValues {
		nextItem, ok := next.Paths.MapOfPathItemValues[path]
		if !ok {
			continue
		}
		for method, baseOp := range baseItem.MapOfOperationValues {
			nextOp, ok := nextItem.MapOfOperationValues[method]
			if !ok {
				continue
			}
			if changes := opChanges(path, method, baseOp, nextOp); len(changes) > 0 {
				r.modified = append(r.modified, changes...)
			}
		}
	}

	return r
}

// pathMethodSet returns a set of "METHOD /path" strings.
func pathMethodSet(spec *openapi3.Spec) map[string]bool {
	m := make(map[string]bool)
	for path, item := range spec.Paths.MapOfPathItemValues {
		for method := range item.MapOfOperationValues {
			m[method+" "+path] = true
		}
	}
	return m
}

// opChanges returns human-readable change descriptions for a single operation.
func opChanges(path, method string, base, next openapi3.Operation) []string {
	var changes []string
	loc := method + " " + path

	baseParams := paramSet(base) // map[key]required
	nextParams := paramSet(next)

	// Parameters removed → breaking if they existed in base.
	for key := range baseParams {
		if _, exists := nextParams[key]; !exists {
			changes = append(changes, fmt.Sprintf("  BREAKING  removed parameter %q from %s", key, loc))
		}
	}
	// New required parameters added → breaking.
	for key, required := range nextParams {
		if _, exists := baseParams[key]; !exists && required {
			changes = append(changes, fmt.Sprintf("  BREAKING  added required parameter %q to %s", key, loc))
		}
	}

	// Response status codes removed → breaking.
	baseStatuses := statusSet(base)
	nextStatuses := statusSet(next)
	for s := range baseStatuses {
		if _, exists := nextStatuses[s]; !exists {
			changes = append(changes, fmt.Sprintf("  BREAKING  removed response status %s from %s", s, loc))
		}
	}
	for s := range nextStatuses {
		if _, exists := baseStatuses[s]; !exists {
			changes = append(changes, fmt.Sprintf("  added     response status %s to %s", s, loc))
		}
	}

	return changes
}

// paramSet returns a map of "in:name" → required for an operation's parameters.
func paramSet(op openapi3.Operation) map[string]bool {
	m := make(map[string]bool, len(op.Parameters))
	for _, p := range op.Parameters {
		if p.Parameter == nil {
			continue
		}
		key := string(p.Parameter.In) + ":" + p.Parameter.Name
		required := p.Parameter.Required != nil && *p.Parameter.Required
		m[key] = required
	}
	return m
}

// statusSet returns a set of HTTP status code strings for an operation.
func statusSet(op openapi3.Operation) map[string]bool {
	m := make(map[string]bool)
	if op.Responses.MapOfResponseOrRefValues == nil {
		return m
	}
	for s := range op.Responses.MapOfResponseOrRefValues {
		m[s] = true
	}
	return m
}

func printDiff(r diffResult) (hasBreaking bool) {
	if len(r.added) == 0 && len(r.removed) == 0 && len(r.modified) == 0 {
		fmt.Println("No differences found.")
		return false
	}

	for _, a := range r.added {
		fmt.Println("  +added    ", a)
	}
	for _, rm := range r.removed {
		fmt.Println("  -BREAKING  removed", rm)
		hasBreaking = true
	}
	for _, m := range r.modified {
		fmt.Println(m)
		if containsBreaking(m) {
			hasBreaking = true
		}
	}
	return hasBreaking
}

func containsBreaking(s string) bool {
	for i := range s {
		if i+8 <= len(s) && s[i:i+8] == "BREAKING" {
			return true
		}
	}
	return false
}
