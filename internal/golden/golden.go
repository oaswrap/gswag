// Package golden provides helpers for golden-file based testing.
//
// Golden files are expected outputs stored under testdata/golden/ relative to the
// test package directory.  When the UPDATE_GOLDEN environment variable is set to
// a non-empty value, the helper writes (or overwrites) golden files with the
// actual output instead of comparing.  Normal test runs compare actual output
// against the stored file and fail with a clear diff when they differ.
//
// Usage:
//
//	go test ./...                    # compare against stored golden files
//	UPDATE_GOLDEN=true go test ./... # regenerate golden files
//	make update-golden               # shorthand via Makefile
package golden

import (
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
)

// update is true when UPDATE_GOLDEN is set to a non-empty value.
// Using an environment variable avoids the need to register a flag in every
// test binary in the module — packages that don't import golden would fail
// with "flag provided but not defined" if a -flag approach were used.
var update = os.Getenv("UPDATE_GOLDEN") != ""

// TB is the subset of testing.TB and ginkgo.FullGinkgoTInterface used by
// Check.  Both *testing.T and the value returned by GinkgoT() satisfy it.
type TB interface {
	Helper()
	Logf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// Check compares actual to the golden file at goldenPath.  When UPDATE_GOLDEN
// is set it writes actual to goldenPath (creating parent directories as needed).
// Otherwise it reads the golden file and fails the test when the contents differ.
func Check(t TB, goldenPath string, actual []byte) {
	t.Helper()

	if update {
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("golden: create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatalf("golden: write %s: %v", goldenPath, err)
		}
		t.Logf("golden: updated %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden: read %s: %v\n(run UPDATE_GOLDEN=true go test ./... to create it)", goldenPath, err)
	}

	if diff := cmp.Diff(string(want), string(actual)); diff != "" {
		t.Fatalf("golden: mismatch for %s (-want +got):\n%s", goldenPath, diff)
	}
}
