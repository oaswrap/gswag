package gswag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/swaggest/openapi-go/openapi3"
)

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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := partialSpecPath(dir, nodeIndex)
	return WriteSpecTo(path, JSON)
}

// MergeAndWriteSpec reads all partial spec files written by WritePartialSpec,
// merges their paths and schemas, then writes the final spec using the global config.
// This must only be called on node 1 after all other nodes have called WritePartialSpec.
func MergeAndWriteSpec(totalNodes int, dir string) error {
	if globalConfig == nil {
		return fmt.Errorf("gswag: not initialised — call Init() first")
	}

	base, err := readPartialSpec(dir, 1)
	if err != nil {
		return fmt.Errorf("gswag: reading node 1 partial: %w", err)
	}

	for i := 2; i <= totalNodes; i++ {
		partial, err := readPartialSpec(dir, i)
		if err != nil {
			return fmt.Errorf("gswag: reading node %d partial: %w", i, err)
		}
		mergeSpec(base, partial)
	}

	// Write merged spec to the configured output path.
	var data []byte
	switch globalConfig.OutputFormat {
	case JSON:
		data, err = json.MarshalIndent(base, "", "  ")
	default: // YAML
		data, err = base.MarshalYAML()
	}
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(globalConfig.OutputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(globalConfig.OutputPath, data, 0o644)
}

func partialSpecPath(dir string, nodeIndex int) string {
	return filepath.Join(dir, fmt.Sprintf("node-%d.json", nodeIndex))
}

func readPartialSpec(dir string, nodeIndex int) (*openapi3.Spec, error) {
	path := partialSpecPath(dir, nodeIndex)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	spec := &openapi3.Spec{}
	if err := json.Unmarshal(data, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// mergeSpec merges paths and schemas from src into dst in place.
func mergeSpec(dst, src *openapi3.Spec) {
	// Merge paths.
	if src.Paths.MapOfPathItemValues == nil {
		return
	}
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
				for method, op := range srcItem.MapOfOperationValues {
					if _, alreadyDefined := existing.MapOfOperationValues[method]; !alreadyDefined {
						existing.MapOfOperationValues[method] = op
					}
				}
			}
			dst.Paths.MapOfPathItemValues[path] = existing
		} else {
			dst.Paths.MapOfPathItemValues[path] = srcItem
		}
	}

	// Merge component schemas.
	if src.Components == nil {
		return
	}
	dst.ComponentsEns()
	if src.Components.Schemas != nil {
		dst.Components.SchemasEns()
		for name, schema := range src.Components.Schemas.MapOfSchemaOrRefValues {
			if _, exists := dst.Components.Schemas.MapOfSchemaOrRefValues[name]; !exists {
				dst.Components.Schemas.WithMapOfSchemaOrRefValuesItem(name, schema)
			}
		}
	}
	if src.Components.SecuritySchemes != nil {
		dst.Components.SecuritySchemesEns()
		for name, ss := range src.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues {
			if _, exists := dst.Components.SecuritySchemes.MapOfSecuritySchemeOrRefValues[name]; !exists {
				dst.Components.SecuritySchemes.WithMapOfSecuritySchemeOrRefValuesItem(name, ss)
			}
		}
	}
}
