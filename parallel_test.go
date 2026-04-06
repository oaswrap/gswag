package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Parallel tests use a separate Init so they don't interfere with the root suite's globalCollector.

func makeEchoServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)
	return srv
}

var _ = Describe("WritePartialSpec", func() {
	It("writes a JSON partial spec file for the current collector", func() {
		dir := GinkgoT().TempDir()

		// Use the current root suite's collector (populated by spec_test.go paths).
		err := gswag.WritePartialSpec(1, dir)
		Expect(err).NotTo(HaveOccurred())

		data, err := os.ReadFile(filepath.Join(dir, "node-1.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(string(data), "Root Suite API")).To(BeTrue())
	})
})

func TestMergeAndWriteSpec(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "merged.yaml")

	srv1 := makeEchoServer(t)
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	gswag.SetTestServer(srv1)

	// Manually trigger DSL registration by calling the spec collector-level helper
	// through a Ginkgo sub-suite would be complex; instead test the merge step directly
	// by writing two partial specs from the filesystem.
	// Write node-1 partial (uses current empty-paths collector — just checks file I/O).
	if err := gswag.WritePartialSpec(1, dir); err != nil {
		t.Fatalf("node 1 WritePartialSpec: %v", err)
	}

	srv2 := makeEchoServer(t)
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})
	gswag.SetTestServer(srv2)

	if err := gswag.WritePartialSpec(2, dir); err != nil {
		t.Fatalf("node 2 WritePartialSpec: %v", err)
	}

	// Re-init with the final output path and merge.
	gswag.Init(&gswag.Config{
		Title:      "Merged API",
		Version:    "1.0.0",
		OutputPath: outPath,
	})

	if err := gswag.MergeAndWriteSpec(2, dir); err != nil {
		t.Fatalf("MergeAndWriteSpec: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading merged spec: %v", err)
	}
	if !strings.Contains(string(data), "Merged API") {
		t.Errorf("expected 'Merged API' in merged spec, got:\n%s", string(data))
	}
}
