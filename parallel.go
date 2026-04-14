package gswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/swaggest/openapi-go/openapi3"
)

const defaultMergeTimeout = 30 * time.Second
const mergePollInterval = 50 * time.Millisecond

// WritePartialSpec serialises the current collector's spec to a file inside dir.
// The file is named after nodeIndex (1-based) so that the merge step can discover
// all node outputs without coordination.
//
// Call this in AfterSuite on every parallel Ginkgo node before shutting down:
//
//	var _ = AfterSuite(func() {
//	    testServer.Close()
//	    Expect(gswag.WritePartialSpec(GinkgoParallelProcess(), "./tmp/gswag")).To(Succeed())
//	    if GinkgoParallelProcess() == 1 {
//	        Expect(gswag.MergeAndWriteSpec(GinkgoProcs(), "./tmp/gswag")).To(Succeed())
//	    }
//	})
func WritePartialSpec(nodeIndex int, dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	path := partialSpecPath(dir, nodeIndex)
	return WriteSpecTo(path, JSON)
}

// MergeAndWriteSpec reads all partial spec files written by WritePartialSpec,
// merges their paths and schemas, then writes the final spec using the global config.
// This must only be called on node 1 after all other nodes have called WritePartialSpec.
//
// It polls for each node's partial file until it appears, using the MergeTimeout
// from the global config (default 30 s). Use Ginkgo's SynchronizedAfterSuite to
// guarantee all nodes have finished writing before this is called.
func MergeAndWriteSpec(totalNodes int, dir string) error {
	if globalConfig == nil {
		return errors.New("gswag: not initialised — call Init() first")
	}

	timeout := globalConfig.MergeTimeout
	if timeout <= 0 {
		timeout = defaultMergeTimeout
	}

	base, err := waitAndReadPartialSpec(dir, 1, timeout)
	if err != nil {
		return fmt.Errorf("gswag: reading node 1 partial: %w", err)
	}

	for i := 2; i <= totalNodes; i++ {
		partial, perr := waitAndReadPartialSpec(dir, i, timeout)
		if perr != nil {
			return fmt.Errorf("gswag: reading node %d partial: %w", i, perr)
		}
		mergeSpec(base, partial)
	}

	// Write merged spec to the configured output path.
	var data []byte
	switch globalConfig.OutputFormat {
	case JSON:
		data, err = json.MarshalIndent(base, "", "  ")
	case YAML:
		data, err = base.MarshalYAML()
	default:
		return fmt.Errorf("unknown output format: %v", globalConfig.OutputFormat)
	}
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(globalConfig.OutputPath), 0o750); err != nil {
		return err
	}
	return os.WriteFile(globalConfig.OutputPath, data, 0o600)
}

func partialSpecPath(dir string, nodeIndex int) string {
	return filepath.Join(dir, fmt.Sprintf("node-%d.json", nodeIndex))
}

// waitAndReadPartialSpec polls for the partial spec file until it appears or timeout expires.
func waitAndReadPartialSpec(dir string, nodeIndex int, timeout time.Duration) (*openapi3.Spec, error) {
	path := partialSpecPath(dir, nodeIndex)
	deadline := time.Now().Add(timeout)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			spec := &openapi3.Spec{}
			if unmarshalErr := json.Unmarshal(data, spec); unmarshalErr != nil {
				return nil, unmarshalErr
			}
			return spec, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for node %d partial spec at %s", nodeIndex, path)
		}
		time.Sleep(mergePollInterval)
	}
}

// mergeSpec merges paths and all component types from src into dst in place.
func mergeSpec(dst, src *openapi3.Spec) {
	// Merge paths (independent of component merging).
	if src.Paths.MapOfPathItemValues != nil {
		if dst.Paths.MapOfPathItemValues == nil {
			dst.Paths.MapOfPathItemValues = make(map[string]openapi3.PathItem)
		}
		for path, srcItem := range src.Paths.MapOfPathItemValues {
			if existing, ok := dst.Paths.MapOfPathItemValues[path]; ok {
				// Merge operations into the existing path item.
				if srcItem.MapOfOperationValues != nil {
					if existing.MapOfOperationValues == nil {
						existing.MapOfOperationValues = make(map[string]openapi3.Operation)
					}
					for method, srcOp := range srcItem.MapOfOperationValues {
						if existingOp, alreadyDefined := existing.MapOfOperationValues[method]; !alreadyDefined {
							existing.MapOfOperationValues[method] = srcOp
						} else {
							// Operation exists in both: merge at the response level so
							// each node contributes the schemas it inferred at runtime.
							mergeOperationResponses(&existingOp, &srcOp)
							existing.MapOfOperationValues[method] = existingOp
						}
					}
				}
				dst.Paths.MapOfPathItemValues[path] = existing
			} else {
				dst.Paths.MapOfPathItemValues[path] = srcItem
			}
		}
	}

	// Merge components — always runs even when src has no paths.
	if src.Components == nil {
		return
	}
	dst.ComponentsEns()

	if src.Components.Schemas != nil {
		dst.Components.SchemasEns()
		for name, v := range src.Components.Schemas.MapOfSchemaOrRefValues {
			if _, exists := dst.Components.Schemas.MapOfSchemaOrRefValues[name]; !exists {
				dst.Components.Schemas.WithMapOfSchemaOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.SecuritySchemes != nil {
		dst.Components.SecuritySchemesEns()
		for name, v := range src.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues {
			if _, exists := dst.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues[name]; !exists {
				dst.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Responses != nil {
		dst.Components.ResponsesEns()
		for name, v := range src.Components.Responses.MapOfResponseOrRefValues {
			if _, exists := dst.Components.Responses.MapOfResponseOrRefValues[name]; !exists {
				dst.Components.Responses.WithMapOfResponseOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Parameters != nil {
		dst.Components.ParametersEns()
		for name, v := range src.Components.Parameters.MapOfParameterOrRefValues {
			if _, exists := dst.Components.Parameters.MapOfParameterOrRefValues[name]; !exists {
				dst.Components.Parameters.WithMapOfParameterOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.RequestBodies != nil {
		dst.Components.RequestBodiesEns()
		for name, v := range src.Components.RequestBodies.MapOfRequestBodyOrRefValues {
			if _, exists := dst.Components.RequestBodies.MapOfRequestBodyOrRefValues[name]; !exists {
				dst.Components.RequestBodies.WithMapOfRequestBodyOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Headers != nil {
		dst.Components.HeadersEns()
		for name, v := range src.Components.Headers.MapOfHeaderOrRefValues {
			if _, exists := dst.Components.Headers.MapOfHeaderOrRefValues[name]; !exists {
				dst.Components.Headers.WithMapOfHeaderOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Examples != nil {
		dst.Components.ExamplesEns()
		for name, v := range src.Components.Examples.MapOfExampleOrRefValues {
			if _, exists := dst.Components.Examples.MapOfExampleOrRefValues[name]; !exists {
				dst.Components.Examples.WithMapOfExampleOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Links != nil {
		dst.Components.LinksEns()
		for name, v := range src.Components.Links.MapOfLinkOrRefValues {
			if _, exists := dst.Components.Links.MapOfLinkOrRefValues[name]; !exists {
				dst.Components.Links.WithMapOfLinkOrRefValuesItem(name, v)
			}
		}
	}
	if src.Components.Callbacks != nil {
		dst.Components.CallbacksEns()
		for name, v := range src.Components.Callbacks.MapOfCallbackOrRefValues {
			if _, exists := dst.Components.Callbacks.MapOfCallbackOrRefValues[name]; !exists {
				dst.Components.Callbacks.WithMapOfCallbackOrRefValuesItem(name, v)
			}
		}
	}
}

// mergeOperationResponses merges response entries from src into dst at the
// status-code level. A response from src is applied when:
//   - dst has no entry for that status code, OR
//   - dst has the status but no content schema (the node ran zero tests for it)
//     while src does have a content schema (inferred from a live response).
//
// It also copies over other runtime-inferred fields (RequestBody) that only
// the node which actually ran the test will have populated.
func mergeOperationResponses(dst, src *openapi3.Operation) {
	// Copy requestBody when the base node never ran this operation's test.
	if dst.RequestBody == nil && src.RequestBody != nil {
		dst.RequestBody = src.RequestBody
	}

	if src.Responses.MapOfResponseOrRefValues == nil {
		return
	}
	for status, srcROR := range src.Responses.MapOfResponseOrRefValues {
		dstROR, exists := dst.Responses.MapOfResponseOrRefValues[status]
		if !exists {
			dst.Responses.WithMapOfResponseOrRefValuesItem(status, srcROR)
			continue
		}
		// Prefer src when dst has no content but src does (e.g. dst came from a
		// node that ran zero tests, src came from a node that inferred the schema).
		if dstROR.Response != nil && len(dstROR.Response.Content) == 0 &&
			srcROR.Response != nil && len(srcROR.Response.Content) > 0 {
			dst.Responses.WithMapOfResponseOrRefValuesItem(status, srcROR)
		}
	}
}
