package gswag

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var data []byte
	var err error

	globalCollector.mu.Lock()
	spec := globalCollector.reflector.Spec
	globalCollector.mu.Unlock()

	switch format {
	case JSON:
		data, err = json.MarshalIndent(spec, "", "  ")
	default: // YAML
		data, err = spec.MarshalYAML()
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
