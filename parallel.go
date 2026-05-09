package gswag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oaswrap/spec/openapi"
)

const defaultMergeTimeout = 30 * time.Second
const mergePollInterval = 50 * time.Millisecond

// WritePartialSpec serialises the current collector's spec to a file inside dir.
func WritePartialSpec(nodeIndex int, dir string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	path := partialSpecPath(dir, nodeIndex)
	return WriteSpecTo(path, JSON)
}

// MergeAndWriteSpec reads all partial spec files, merges them, then writes the final spec.
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
	var data []byte
	switch globalConfig.OutputFormat {
	case JSON:
		raw, marshalErr := openapi.MarshalJSON(base)
		if marshalErr != nil {
			return marshalErr
		}
		data, err = json.MarshalIndent(json.RawMessage(raw), "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	case YAML:
		data, err = openapi.MarshalYAML(base)
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
func waitAndReadPartialSpec(dir string, nodeIndex int, timeout time.Duration) (*openapi.Document, error) {
	path := partialSpecPath(dir, nodeIndex)
	deadline := time.Now().Add(timeout)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			specDoc := &openapi.Document{}
			if unmarshalErr := json.Unmarshal(data, specDoc); unmarshalErr != nil {
				return nil, unmarshalErr
			}
			return specDoc, nil
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
func mergeSpec(dst, src *openapi.Document) {
	if dst == nil || src == nil {
		return
	}
	if src.Paths != nil {
		if dst.Paths == nil {
			dst.Paths = map[string]*openapi.PathItem{}
		}
		for path, srcItem := range src.Paths {
			if srcItem == nil {
				continue
			}
			if existing := dst.Paths[path]; existing != nil {
				mergePathItem(existing, srcItem)
			} else {
				dst.Paths[path] = srcItem
			}
		}
	}
	mergeComponents(dst, src)
}

func mergePathItem(dst, src *openapi.PathItem) {
	mergeOperationSlot(&dst.Get, src.Get)
	mergeOperationSlot(&dst.Put, src.Put)
	mergeOperationSlot(&dst.Post, src.Post)
	mergeOperationSlot(&dst.Delete, src.Delete)
	mergeOperationSlot(&dst.Options, src.Options)
	mergeOperationSlot(&dst.Head, src.Head)
	mergeOperationSlot(&dst.Patch, src.Patch)
	mergeOperationSlot(&dst.Trace, src.Trace)
	mergeOperationSlot(&dst.Query, src.Query)
	if len(src.AdditionalOperations) > 0 {
		if dst.AdditionalOperations == nil {
			dst.AdditionalOperations = map[string]*openapi.Operation{}
		}
		for method, srcOp := range src.AdditionalOperations {
			if dstOp := dst.AdditionalOperations[method]; dstOp != nil {
				mergeOperationResponses(dstOp, srcOp)
			} else {
				dst.AdditionalOperations[method] = srcOp
			}
		}
	}
}

func mergeOperationSlot(dst **openapi.Operation, src *openapi.Operation) {
	if src == nil {
		return
	}
	if *dst == nil {
		*dst = src
		return
	}
	mergeOperationResponses(*dst, src)
}

func mergeComponents(dst, src *openapi.Document) {
	if src.Components == nil {
		return
	}
	if dst.Components == nil {
		dst.Components = &openapi.Components{}
	}
	if len(src.Components.Schemas) > 0 {
		if dst.Components.Schemas == nil {
			dst.Components.Schemas = map[string]*openapi.Schema{}
		}
		for name, v := range src.Components.Schemas {
			if _, exists := dst.Components.Schemas[name]; !exists {
				dst.Components.Schemas[name] = v
			}
		}
	}
	if len(src.Components.SecuritySchemes) > 0 {
		if dst.Components.SecuritySchemes == nil {
			dst.Components.SecuritySchemes = map[string]*openapi.SecurityScheme{}
		}
		for name, v := range src.Components.SecuritySchemes {
			if _, exists := dst.Components.SecuritySchemes[name]; !exists {
				dst.Components.SecuritySchemes[name] = v
			}
		}
	}
	if len(src.Components.Responses) > 0 {
		if dst.Components.Responses == nil {
			dst.Components.Responses = map[string]*openapi.Response{}
		}
		for name, v := range src.Components.Responses {
			if _, exists := dst.Components.Responses[name]; !exists {
				dst.Components.Responses[name] = v
			}
		}
	}
	if len(src.Components.Parameters) > 0 {
		if dst.Components.Parameters == nil {
			dst.Components.Parameters = map[string]*openapi.Parameter{}
		}
		for name, v := range src.Components.Parameters {
			if _, exists := dst.Components.Parameters[name]; !exists {
				dst.Components.Parameters[name] = v
			}
		}
	}
	if len(src.Components.RequestBodies) > 0 {
		if dst.Components.RequestBodies == nil {
			dst.Components.RequestBodies = map[string]*openapi.RequestBody{}
		}
		for name, v := range src.Components.RequestBodies {
			if _, exists := dst.Components.RequestBodies[name]; !exists {
				dst.Components.RequestBodies[name] = v
			}
		}
	}
	if len(src.Components.Headers) > 0 {
		if dst.Components.Headers == nil {
			dst.Components.Headers = map[string]*openapi.Header{}
		}
		for name, v := range src.Components.Headers {
			if _, exists := dst.Components.Headers[name]; !exists {
				dst.Components.Headers[name] = v
			}
		}
	}
	if len(src.Components.Examples) > 0 {
		if dst.Components.Examples == nil {
			dst.Components.Examples = map[string]*openapi.Example{}
		}
		for name, v := range src.Components.Examples {
			if _, exists := dst.Components.Examples[name]; !exists {
				dst.Components.Examples[name] = v
			}
		}
	}
	if len(src.Components.Links) > 0 {
		if dst.Components.Links == nil {
			dst.Components.Links = map[string]*openapi.Link{}
		}
		for name, v := range src.Components.Links {
			if _, exists := dst.Components.Links[name]; !exists {
				dst.Components.Links[name] = v
			}
		}
	}
	if len(src.Components.Callbacks) > 0 {
		if dst.Components.Callbacks == nil {
			dst.Components.Callbacks = map[string]*openapi.Callback{}
		}
		for name, v := range src.Components.Callbacks {
			if _, exists := dst.Components.Callbacks[name]; !exists {
				dst.Components.Callbacks[name] = v
			}
		}
	}
}

// mergeOperationResponses merges response entries from src into dst at the status-code level.
func mergeOperationResponses(dst, src *openapi.Operation) {
	if dst == nil || src == nil {
		return
	}
	if dst.RequestBody == nil && src.RequestBody != nil {
		dst.RequestBody = src.RequestBody
	}
	if dst.Responses == nil {
		dst.Responses = map[string]*openapi.Response{}
	}
	for status, srcResp := range src.Responses {
		dstResp := dst.Responses[status]
		if dstResp == nil {
			dst.Responses[status] = srcResp
			continue
		}
		if len(dstResp.Content) == 0 && srcResp != nil && len(srcResp.Content) > 0 {
			dst.Responses[status] = srcResp
		}
	}
}
