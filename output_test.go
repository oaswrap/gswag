package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
)

func initAndRecord(t *testing.T, outputPath string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1}`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)

	gswag.Init(&gswag.Config{
		Title:      "Output Test API",
		Version:    "1.0.0",
		OutputPath: outputPath,
	})
	gswag.GET("/items").WithTag("items").WithSummary("List").Do(srv)
}

func TestWriteSpec_YAML(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "openapi.yaml")
	initAndRecord(t, outPath)

	if err := gswag.WriteSpec(); err != nil {
		t.Fatalf("WriteSpec failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if !strings.Contains(string(data), "Output Test API") {
		t.Errorf("expected title in YAML output, got:\n%s", string(data))
	}
}

func TestWriteSpecTo_JSON(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "openapi.json")
	initAndRecord(t, outPath)

	if err := gswag.WriteSpecTo(outPath, gswag.JSON); err != nil {
		t.Fatalf("WriteSpecTo JSON failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if !strings.Contains(string(data), `"Output Test API"`) {
		t.Errorf("expected title in JSON output, got:\n%s", string(data))
	}
}

func TestWriteSpecTo_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nested", "deep", "openapi.yaml")
	initAndRecord(t, outPath)

	if err := gswag.WriteSpecTo(outPath, gswag.YAML); err != nil {
		t.Fatalf("WriteSpecTo failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected file at %s: %v", outPath, err)
	}
}
