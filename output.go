package gswag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	outputpkg "github.com/oaswrap/gswag/internal/output"
	"github.com/oaswrap/spec/openapi"
)

// WriteSpec serialises the collected spec to the path and format configured via Init.
func WriteSpec() error {
	if globalCollector == nil {
		return nil
	}
	return WriteSpecTo(globalConfig.OutputPath, globalConfig.OutputFormat)
}

// WriteSpecTo serialises the collected spec to a specific path and format.
func WriteSpecTo(path string, format OutputFormat) error {
	if globalCollector == nil {
		return nil
	}
	flushPendingDSLOps()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	globalCollector.mu.Lock()
	outputpkg.SanitizeSpecForSerialization(globalCollector.doc)
	specDoc := globalCollector.doc
	globalCollector.mu.Unlock()

	var data []byte
	var err error
	switch format {
	case JSON:
		raw, marshalErr := openapi.MarshalJSON(specDoc)
		if marshalErr != nil {
			return marshalErr
		}
		data, err = json.MarshalIndent(json.RawMessage(raw), "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	case YAML:
		data, err = openapi.MarshalYAML(specDoc)
	default:
		return fmt.Errorf("unknown output format: %v", format)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
